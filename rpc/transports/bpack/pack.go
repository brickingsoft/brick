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
	"github.com/quic-go/quic-go/quicvarint"
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

func New(options ...Option) (pack *Pack, err error) {
	opts := &Options{
		maxHeaderSize: DefaultMaxHeaderSize,
		fields:        nil,
	}
	for _, option := range options {
		if err = option(opts); err != nil {
			return
		}
	}
	pack = &Pack{
		maxHeaderBytes: opts.maxHeaderSize,
		dict:           new(Dictionary),
	}
	pack.dict.Load(opts.fields)
	return
}

type Pack struct {
	maxHeaderBytes int
	dict           *Dictionary
}

type HeaderIterator iter.Seq2[[]byte, []byte]

func (pack *Pack) EncodeTo(w io.Writer, iter HeaderIterator) (err error) {
	if w == nil {
		err = errors.New("bpack encode to nil Writer")
		return
	}
	if iter == nil {
		err = errors.New("bpack encode to nil HeaderIterator")
		return
	}
	buf := bytebuffers.Acquire()
	defer bytebuffers.Release(buf)

	for name, value := range iter {
		if len(name) == 0 {
			err = errors.New("bpack encode to empty name")
			return
		}
		if len(value) == 0 {
			err = errors.New("bpack encode to empty value")
			return
		}
		i, j := pack.dict.Index(name, value)
		if i < 0 { // not found
			pack.writeLiteralFieldWithoutNameReference(buf, name, value)
		} else if j < 0 { // found name
			pack.writeLiteralFieldWithNameReference(buf, i, value)
		} else { // found name and value
			pack.writeIndexedField(buf, j)
		}
	}
	bLen := buf.Len()
	if bLen > pack.maxHeaderBytes {
		err = errors.New("bpack encode to larger than maxHeaderBytes")
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

func (pack *Pack) writeLiteral(b bytebuffers.Buffer, p []byte) {
	if len(p) == 0 {
		np := quicvarint.Append(nil, uint64(0))
		_, _ = b.Write(np)
		return
	}
	s := unsafe.String(unsafe.SliceData(p), len(p))
	sp := hpack.AppendHuffmanString(nil, s)
	np := quicvarint.Append(nil, uint64(len(sp)))
	_, _ = b.Write(np)
	_, _ = b.Write(sp)
}

func (pack *Pack) writeLiteralFieldWithoutNameReference(b bytebuffers.Buffer, name []byte, value []byte) {
	_ = b.WriteByte(literalFieldWithoutNameReference)
	pack.writeLiteral(b, name)
	pack.writeLiteral(b, value)
}

func (pack *Pack) writeLiteralFieldWithNameReference(b bytebuffers.Buffer, i int, value []byte) {
	_ = b.WriteByte(literalFieldWithNameReference)
	p := quicvarint.Append(nil, uint64(i))
	_, _ = b.Write(p)
	pack.writeLiteral(b, value)
}

func (pack *Pack) writeIndexedField(b bytebuffers.Buffer, i int) {
	_ = b.WriteByte(indexField)
	p := quicvarint.Append(nil, uint64(i))
	_, _ = b.Write(p)
}

type HeaderWriter interface {
	Set(key []byte, value []byte)
}

func (pack *Pack) DecodeFrom(r io.Reader, header HeaderWriter) (err error) {
	if header == nil {
		err = errors.New("bpack decode to nil Header")
		return
	}
	buf := bytebuffers.Acquire()
	defer bytebuffers.Release(buf)

	hn, rhErr := buf.ReadFromLimited(r, 2)
	if hn != 2 {
		if rhErr != nil {
			err = rhErr
			return
		}
		err = errors.New("bpack decode to invalid header length size")
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
			err = rbErr
			return
		}
		err = errors.New("bpack decode to invalid header body size")
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
			err = pack.readIndexedField(buf, header)
			break
		case literalFieldWithNameReference:
			err = pack.readLiteralFieldWithNameReference(buf, header)
			break
		case literalFieldWithoutNameReference:
			err = pack.readLiteralFieldWithoutNameReference(buf, header)
			break
		default:
			err = errors.New("bpack decode to header failed for unknown field kind")
			break
		}
	}

	if errors.Is(err, io.EOF) {
		err = nil
	}
	return
}

func (pack *Pack) readLiteral(b bytebuffers.Buffer) (p []byte, err error) {
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

func (pack *Pack) readLiteralFieldWithoutNameReference(b bytebuffers.Buffer, header HeaderWriter) (err error) {
	name, nameErr := pack.readLiteral(b)
	if nameErr != nil {
		err = nameErr
		return
	}
	value, valueErr := pack.readLiteral(b)
	if valueErr != nil {
		err = valueErr
		return
	}
	header.Set(name, value)
	return
}

func (pack *Pack) readLiteralFieldWithNameReference(b bytebuffers.Buffer, header HeaderWriter) (err error) {
	ni, niErr := quicvarint.Read(b)
	if niErr != nil {
		err = niErr
		return
	}
	name, _ := pack.dict.Get(int(ni))
	if len(name) == 0 {
		err = errors.New("bpack decode to header failed for dict unmatched")
		return
	}
	value, valueErr := pack.readLiteral(b)
	if valueErr != nil {
		err = valueErr
		return
	}
	header.Set(name, value)
	return
}

func (pack *Pack) readIndexedField(b bytebuffers.Buffer, header HeaderWriter) (err error) {
	ni, niErr := quicvarint.Read(b)
	if niErr != nil {
		err = niErr
		return
	}
	name, value := pack.dict.Get(int(ni))
	if len(name) == 0 || len(value) == 0 {
		err = errors.New("bpack decode to header failed for dict unmatched")
		return
	}
	header.Set(name, value)
	return
}

func (pack *Pack) Reset(fields []HeaderField) {
	pack.dict.Load(fields)
}

func (pack *Pack) DumpTo(w io.Writer) (err error) {
	buf := bytebuffers.Acquire()
	defer bytebuffers.Release(buf)

	p, bErr := buf.Borrow(2)
	if bErr != nil {
		err = bErr
		return
	}
	binary.LittleEndian.PutUint16(p, uint16(pack.maxHeaderBytes))
	buf.Return(2)

	pack.dict.Range(func(_ int, name []byte, value []byte) bool {
		pack.writeLiteral(buf, name)
		pack.writeLiteral(buf, value)
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

func (pack *Pack) LoadFrom(r io.Reader) (err error) {
	buf := bytebuffers.Acquire()
	defer bytebuffers.Release(buf)

	hn, rhErr := buf.ReadFromLimited(r, 4)
	if rhErr != nil {
		err = rhErr
		return
	}
	if hn != 4 {
		err = errors.New("bpack load from invalid size")
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
		err = errors.New("bpack load from invalid size")
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
		name, nameErr := pack.readLiteral(buf)
		if nameErr != nil {
			err = nameErr
			break
		}
		value, valueErr := pack.readLiteral(buf)
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

	pack.maxHeaderBytes = int(maxHeaderSize)
	pack.dict.Load(hw.fields)
	return
}
