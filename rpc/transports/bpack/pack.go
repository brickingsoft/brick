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
	DefaultMaxHeaderSize = 8 * 1024
)

type Options struct {
	maxHeaderSize int
	fields        []HeaderField
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
	defer buf.Discard(buf.Len())

	sp, spErr := buf.Borrow(2)
	if spErr != nil {
		err = spErr
		return
	}
	buf.Return(2)
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
	binary.LittleEndian.PutUint16(sp, uint16(bLen))
	_, err = buf.WriteTo(w)
	return
}

func (pack *Pack) writeLiteral(b bytebuffers.Buffer, p []byte) {
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

	h, hErr := buf.Borrow(2)
	if hErr != nil {
		err = hErr
		return
	}
	buf.Return(2)
	_, lnRErr := io.ReadFull(r, h)
	if lnRErr != nil {
		err = lnRErr
		return
	}
	bLen := binary.LittleEndian.Uint16(h)
	buf.Discard(2)

	if bLen == 0 {
		return
	}

	b, bErr := buf.Borrow(int(bLen))
	if bErr != nil {
		err = bErr
		return
	}
	buf.Return(int(bLen))

	_, rbErr := io.ReadFull(r, b)
	if rbErr != nil {
		err = rbErr
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

	return
}

func (pack *Pack) readLiteral(b bytebuffers.Buffer) (p []byte, err error) {
	n, nErr := quicvarint.Read(b)
	if nErr != nil {
		err = nErr
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

	h, hErr := buf.Borrow(10)
	if hErr != nil {
		err = hErr
		return
	}
	buf.Return(10)
	binary.LittleEndian.PutUint16(h[:2], uint16(pack.maxHeaderBytes))

	pack.dict.Range(func(_ int, name []byte, value []byte) bool {
		pack.writeLiteralFieldWithoutNameReference(buf, name, value)
		return true
	})

	n := buf.Len() - 10
	binary.LittleEndian.PutUint64(h[2:10], uint64(n))
	_, err = buf.WriteTo(w)
	return
}

func (pack *Pack) LoadFrom(r io.Reader) (err error) {
	buf := bytebuffers.Acquire()
	defer bytebuffers.Release(buf)

	h, hErr := buf.Borrow(10)
	if hErr != nil {
		err = hErr
		return
	}
	buf.Return(10)
	_, rhErr := io.ReadFull(r, h)
	if rhErr != nil {
		err = rhErr
		return
	}
	maxHeaderSize := binary.LittleEndian.Uint16(h[:2])

	bLen := binary.LittleEndian.Uint64(h[2:10])
	if bLen == 0 {
		return
	}
	buf.Discard(10)
	b, bErr := buf.Borrow(int(bLen))
	if bErr != nil {
		err = bErr
		return
	}
	buf.Return(int(bLen))
	_, rbErr := io.ReadFull(r, b)
	if rbErr != nil {
		err = rbErr
		return
	}

	hw := new(HeaderFields)

	if err = pack.readLiteralFieldWithoutNameReference(buf, hw); err != nil {
		return
	}

	pack.maxHeaderBytes = int(maxHeaderSize)
	pack.dict.Load(hw.fields)
	return
}
