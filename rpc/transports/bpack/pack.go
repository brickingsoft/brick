package bpack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"iter"
	"math"
	"slices"
	"strings"
	"unsafe"

	"github.com/brickingsoft/brick/pkg/quicvarint"
	"github.com/brickingsoft/bytebuffers"
	"golang.org/x/net/http2/hpack"
)

const (
	literalFieldWithoutNameReference byte = 0x01
	literalFieldWithNameReference    byte = 0x02
	indexField                       byte = 0x03
)

const (
	DefaultMaxHeaderSize      = 8 * 1024
	maxHeaderFieldContentSize = 32 * 1024 * 1024
)

type Options struct {
	maxHeaderSize             int
	maxHeaderFieldContentSize int
	fields                    []HeaderField
}

type Option func(*Options) error

func MaxHeaderSize(n int) Option {
	return func(o *Options) error {
		if n < 64 {
			return errors.New("max header size must be at least 64")
		}
		if n > math.MaxUint16 {
			return errors.New("max header size must be less than or equal to 63kb")
		}
		o.maxHeaderSize = n
		return nil
	}
}

func Field(name string, values ...string) Option {
	return func(o *Options) (err error) {
		name = strings.ToLower(strings.TrimSpace(name))
		if name == "" {
			return nil
		}
		exist := false
		for _, field := range o.fields {
			if field.Name == name {
				exist = true
				break
			}
		}
		o.maxHeaderFieldContentSize += len(name)
		if o.maxHeaderFieldContentSize > maxHeaderFieldContentSize {
			return errors.New("max header field content size exceeds limit")
		}
		o.fields = append(o.fields, HeaderField{Name: name})
		if len(values) == 0 {
			if exist {
				return
			}
			return
		}
		for _, value := range values {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			vp := []byte(value)
			if slices.ContainsFunc(o.fields, func(hf HeaderField) bool {
				return hf.Name == name && bytes.Equal(hf.Value, vp)
			}) {
				continue
			}
			o.maxHeaderFieldContentSize += len(name)
			o.maxHeaderFieldContentSize += len(vp)
			if o.maxHeaderFieldContentSize > maxHeaderFieldContentSize {
				return errors.New("max header field content size exceeds limit")
			}
			o.fields = append(o.fields, HeaderField{Name: name, Value: vp})
		}
		return
	}
}

func New(options ...Option) (pack *Packer, err error) {
	opts := &Options{
		maxHeaderSize: DefaultMaxHeaderSize,
		fields:        nil,
	}
	for _, option := range options {
		if err = option(opts); err != nil {
			return
		}
	}
	pack = &Packer{
		maxHeaderBytes: opts.maxHeaderSize,
		dict:           new(Dictionary),
	}
	pack.dict.Load(opts.fields)
	return
}

type Packer struct {
	maxHeaderBytes int
	dict           *Dictionary
}

type HeaderIterator iter.Seq2[[]byte, []byte]

func (packer *Packer) PackTo(w io.Writer, iter HeaderIterator) (err error) {
	if w == nil {
		err = errors.New("pack failed for writer is nil")
		return
	}
	if iter == nil {
		err = errors.New("pack failed for header iterator is nil")
		return
	}
	buf := bytebuffers.Acquire()
	defer bytebuffers.Release(buf)

	for name, value := range iter {
		if len(name) == 0 {
			err = errors.New("pack failed for one empty name")
			return
		}
		if len(value) == 0 {
			err = errors.New("pack failed for one empty value")
			return
		}
		i, j := packer.dict.Index(name, value)
		if i < 0 { // not found
			packer.writeLiteralFieldWithoutNameReference(buf, name, value)
		} else if j < 0 { // found name
			packer.writeLiteralFieldWithNameReference(buf, i, value)
		} else { // found name and value
			packer.writeIndexedField(buf, j)
		}
	}
	bLen := buf.Len()
	if bLen > packer.maxHeaderBytes {
		err = errors.New("pack to larger than maxHeaderBytes")
		return
	}

	spp := acquireBytes()
	defer releaseBytes(spp)
	sp := (*spp)[:2]
	binary.LittleEndian.PutUint16(sp, uint16(bLen))
	for wn := 0; wn < 2; {
		n, wErr := w.Write(sp[wn:])
		if wErr != nil {
			err = wErr
			return
		}
		wn += n
	}

	_, err = buf.WriteTo(w)
	return
}

func (packer *Packer) writeLiteral(b bytebuffers.Buffer, p []byte) {

	pLen := uint64(len(p))
	if pLen == 0 {
		_, _ = quicvarint.Write(b, pLen)
		return
	}
	s := unsafe.String(unsafe.SliceData(p), pLen)
	if hpack.HuffmanEncodeLength(s) < pLen {
		sp := hpack.AppendHuffmanString(nil, s)
		_, _ = quicvarint.Write(b, uint64(len(sp)))
		_, _ = b.Write(sp)
	} else {
		_, _ = quicvarint.Write(b, pLen)
		_, _ = b.Write(p)
	}
}

func (packer *Packer) writeLiteralFieldWithoutNameReference(b bytebuffers.Buffer, name []byte, value []byte) {
	_ = b.WriteByte(literalFieldWithoutNameReference)
	packer.writeLiteral(b, name)
	packer.writeLiteral(b, value)
}

func (packer *Packer) writeLiteralFieldWithNameReference(b bytebuffers.Buffer, i int, value []byte) {
	_ = b.WriteByte(literalFieldWithNameReference)
	_, _ = quicvarint.Write(b, uint64(i))
	packer.writeLiteral(b, value)
}

func (packer *Packer) writeIndexedField(b bytebuffers.Buffer, i int) {
	_ = b.WriteByte(indexField)
	_, _ = quicvarint.Write(b, uint64(i))
}

type HeaderWriter interface {
	Set(key []byte, value []byte) error
}

func (packer *Packer) UnpackFrom(r io.Reader, header HeaderWriter) (err error) {
	if r == nil {
		err = errors.New("unpack failed for reader is nil")
		return
	}
	if header == nil {
		err = errors.New("unpack failed for header writer is nil")
		return
	}
	buf := bytebuffers.Acquire()
	defer bytebuffers.Release(buf)

	hn, rhErr := buf.ReadFromLimited(r, 2)
	if hn != 2 {
		if rhErr != nil {
			err = errors.Join(errors.New("unpack failed"), rhErr)
			return
		}
		err = errors.New("unpack failed for invalid header length size")
		return
	}
	h, hErr := buf.Next(2)
	if hErr != nil {
		err = hErr
		return
	}
	bLen := binary.LittleEndian.Uint16(h)
	if bLen == 0 {
		return
	}

	bn, rbErr := buf.ReadFromLimited(r, int(bLen))
	if bn != int(bLen) {
		if rbErr != nil {
			err = errors.Join(errors.New("unpack failed"), rbErr)
			return
		}
		err = errors.New("unpack failed for invalid header body size")
		return
	}

	for err == nil {
		k, kErr := buf.ReadByte()
		if kErr != nil {
			err = kErr
			break
		}
		switch k {
		case indexField:
			err = packer.readIndexedField(buf, header)
			break
		case literalFieldWithNameReference:
			err = packer.readLiteralFieldWithNameReference(buf, header)
			break
		case literalFieldWithoutNameReference:
			err = packer.readLiteralFieldWithoutNameReference(buf, header)
			break
		default:
			err = errors.New("unknown field kind")
			break
		}
	}

	if errors.Is(err, io.EOF) {
		err = nil
	} else {
		errors.Join(errors.New("unpack failed"), err)
	}
	return
}

func (packer *Packer) readLiteral(b bytebuffers.Buffer) (p []byte, err error) {
	n, nErr := quicvarint.Read(b)
	if nErr != nil {
		err = nErr
		return
	}
	if n == 0 {
		return
	}
	p, err = b.Next(int(n))
	if err != nil {
		return
	}
	if usesHuffman := p[0]&0x80 > 0; usesHuffman {
		s, sErr := hpack.HuffmanDecodeToString(p)
		if sErr != nil {
			err = sErr
			return
		}
		p = []byte(s)
	}
	return
}

func (packer *Packer) readLiteralFieldWithoutNameReference(b bytebuffers.Buffer, header HeaderWriter) (err error) {
	name, nameErr := packer.readLiteral(b)
	if nameErr != nil {
		err = nameErr
		return
	}
	value, valueErr := packer.readLiteral(b)
	if valueErr != nil {
		err = valueErr
		return
	}
	err = header.Set(name, value)
	return
}

func (packer *Packer) readLiteralFieldWithNameReference(b bytebuffers.Buffer, header HeaderWriter) (err error) {
	ni, niErr := quicvarint.Read(b)
	if niErr != nil {
		err = niErr
		return
	}
	name, _ := packer.dict.Get(int(ni))
	if len(name) == 0 {
		err = errors.New("read no name from dictionary")
		return
	}
	value, valueErr := packer.readLiteral(b)
	if valueErr != nil {
		err = valueErr
		return
	}
	err = header.Set(name, value)
	return
}

func (packer *Packer) readIndexedField(b bytebuffers.Buffer, header HeaderWriter) (err error) {
	ni, niErr := quicvarint.Read(b)
	if niErr != nil {
		err = niErr
		return
	}
	name, value := packer.dict.Get(int(ni))
	if len(name) == 0 || len(value) == 0 {
		err = errors.New("read nothing from dictionary")
		return
	}
	err = header.Set(name, value)
	return
}

func (packer *Packer) Reset(fields []HeaderField) {
	packer.dict.Load(fields)
}

func (packer *Packer) DumpTo(w io.Writer) (err error) {
	buf := bytebuffers.Acquire()
	defer bytebuffers.Release(buf)

	p, bErr := buf.Borrow(2)
	if bErr != nil {
		err = bErr
		return
	}
	binary.LittleEndian.PutUint16(p, uint16(packer.maxHeaderBytes))
	buf.Return(2)

	packer.dict.Range(func(_ int, name []byte, value []byte) bool {
		packer.writeLiteral(buf, name)
		packer.writeLiteral(buf, value)
		return true
	})

	n := buf.Len()
	spp := acquireBytes()
	defer releaseBytes(spp)
	sp := (*spp)[:4]
	binary.LittleEndian.PutUint32(sp, uint32(n))

	for wn := 0; wn < 4; {
		nn, wErr := w.Write(sp[wn:])
		if wErr != nil {
			err = wErr
			return
		}
		wn += nn
	}
	_, err = buf.WriteTo(w)
	return
}

func (packer *Packer) LoadFrom(r io.Reader) (err error) {
	buf := bytebuffers.Acquire()
	defer bytebuffers.Release(buf)

	hn, rhErr := buf.ReadFromLimited(r, 4)
	if rhErr != nil {
		err = rhErr
		return
	}
	if hn != 4 {
		err = errors.New("packer load from invalid size")
		return
	}
	h, hErr := buf.Next(4)
	if hErr != nil {
		err = hErr
		return
	}
	bLen := binary.LittleEndian.Uint32(h[:4])

	if bLen == 0 {
		return
	}

	bn, rbErr := buf.ReadFromLimited(r, int(bLen))
	if bn != int(bLen) {
		if rbErr != nil {
			err = rbErr
			return
		}
		err = errors.New("packer load from invalid size")
		return
	}
	mp, mpErr := buf.Next(2)
	if mpErr != nil {
		err = mpErr
		return
	}
	maxHeaderSize := binary.LittleEndian.Uint16(mp)
	bLen -= 2

	hw := new(HeaderFields)
	for {
		name, nameErr := packer.readLiteral(buf)
		if nameErr != nil {
			err = nameErr
			break
		}
		value, valueErr := packer.readLiteral(buf)
		if valueErr != nil {
			err = valueErr
			break
		}
		if len(name) == 0 {
			continue
		}
		hw.Set(name, value)
	}

	if errors.Is(err, io.EOF) {
		err = nil
	} else {
		return
	}

	packer.maxHeaderBytes = int(maxHeaderSize)
	packer.dict.Load(hw.fields)
	return
}
