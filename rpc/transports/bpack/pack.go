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

	"github.com/brickingsoft/bytebuffers"
	"golang.org/x/net/http2/hpack"
)

const (
	DefaultMaxHeaderSize = 8 * 1024
	maxHeaderSize        = math.MaxUint16 - 2
)

type Options struct {
	maxHeaderSize int
	fields        []HeaderField
}

type Option func(*Options) error

func MaxHeaderSize(n int) Option {
	return func(o *Options) error {
		if n < 1 {
			n = DefaultMaxHeaderSize
		}
		if n > maxHeaderSize {
			return errors.New("max header size must be less than to 63kb")
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

func (packer *Packer) writeLiteral(b bytebuffers.Buffer, prefix byte, n byte, p []byte) {
	pLen := uint64(len(p))
	if pLen == 0 {
		writeVarInt(b, prefix, n, pLen)
		return
	}
	s := unsafe.String(unsafe.SliceData(p), pLen)
	if hpack.HuffmanEncodeLength(s) < pLen {
		sp := hpack.AppendHuffmanString(nil, s)
		writeVarInt(b, prefix, n, uint64(len(sp)))
		_, _ = b.Write(sp)
	} else {
		writeVarInt(b, prefix, n, pLen)
		_, _ = b.Write(p)
	}
}

func (packer *Packer) writeLiteralFieldWithoutNameReference(b bytebuffers.Buffer, name []byte, value []byte) {
	packer.writeLiteral(b, 0x20^0x8, 3, name)
	packer.writeLiteral(b, 0x80, 7, value)
}

func (packer *Packer) writeLiteralFieldWithNameReference(b bytebuffers.Buffer, i int, value []byte) {
	writeVarInt(b, 0x50, 4, uint64(i))
	packer.writeLiteral(b, 0x80, 7, value)
}

func (packer *Packer) writeIndexedField(b bytebuffers.Buffer, i int) {
	writeVarInt(b, 0xc0, 6, uint64(i))
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

	h := buf.Peek(2)
	bLen := binary.LittleEndian.Uint16(h)
	buf.Discard(2)
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

	for buf.Len() > 0 && err == nil {
		prefix := buf.Peek(1)[0]
		switch {
		case prefix&0x80 > 0:
			err = packer.readIndexedField(buf, header)
			break
		case prefix&0xc0 == 0x40:
			err = packer.readLiteralFieldWithNameReference(buf, header)
			break
		case prefix&0xe0 == 0x20:
			err = packer.readLiteralFieldWithoutNameReference(buf, header)
			break
		default:
			err = errors.New("unknown field kind")
			return
		}
	}
	if errors.Is(err, io.EOF) {
		err = nil
	} else {
		errors.Join(errors.New("unpack failed"), err)
	}
	return
}

func (packer *Packer) readLiteral(b bytebuffers.Buffer, n byte) (p []byte, err error) {
	i, iErr := readVarInt(b, n)
	if iErr != nil {
		err = iErr
		return
	}
	if i == 0 {
		return
	}
	p, err = b.Next(int(i))
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
	name, nameErr := packer.readLiteral(b, 3)
	if nameErr != nil {
		err = nameErr
		return
	}
	value, valueErr := packer.readLiteral(b, 7)
	if valueErr != nil {
		err = valueErr
		return
	}
	err = header.Set(name, value)
	return
}

func (packer *Packer) readLiteralFieldWithNameReference(b bytebuffers.Buffer, header HeaderWriter) (err error) {
	ni, niErr := readVarInt(b, 4)
	if niErr != nil {
		err = niErr
		return
	}
	name, _ := packer.dict.Get(int(ni))
	if len(name) == 0 {
		err = errors.New("read no name from dictionary")
		return
	}
	value, valueErr := packer.readLiteral(b, 7)
	if valueErr != nil {
		err = valueErr
		return
	}
	err = header.Set(name, value)
	return
}

func (packer *Packer) readIndexedField(b bytebuffers.Buffer, header HeaderWriter) (err error) {
	ni, niErr := readVarInt(b, 6)
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
		packer.writeLiteralFieldWithoutNameReference(buf, name, value)
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
	maxSize := binary.LittleEndian.Uint16(mp)
	bLen -= 2

	hw := new(HeaderFields)
	for {
		if err = packer.readLiteralFieldWithoutNameReference(buf, hw); err != nil {
			break
		}
	}

	if errors.Is(err, io.EOF) {
		err = nil
	} else {
		return
	}

	packer.maxHeaderBytes = int(maxSize)
	packer.dict.Load(hw.fields)
	return
}
