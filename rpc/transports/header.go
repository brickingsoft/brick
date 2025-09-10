package transports

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"sort"
	"strings"
	"sync"
	"unsafe"

	"github.com/brickingsoft/brick/pkg/quicvarint"
	"github.com/brickingsoft/brick/rpc/transports/bpack"
)

var (
	AgentHeaderKey                 = []byte("agent")
	AgentHeaderStringKey           = string(AgentHeaderKey)
	ForwardedHeaderKey             = []byte("forwarded")
	ForwardedHeaderStringKey       = string(ForwardedHeaderKey)
	AuthorizationHeaderKey         = []byte("authorization")
	AuthorizationHeaderStringKey   = string(AuthorizationHeaderKey)
	ContentLengthHeaderKey         = []byte("content-length")
	ContentLengthHeaderStringKey   = string(ContentLengthHeaderKey)
	ContentTypeHeaderKey           = []byte("content-type")
	ContentTypeHeaderStringKey     = string(ContentTypeHeaderKey)
	ContentEncodingHeaderKey       = []byte("content-encoding")
	ContentEncodingHeaderStringKey = string(ContentEncodingHeaderKey)
	SignatureHeaderKey             = []byte("signature")
	SignatureHeaderStringKey       = string(SignatureHeaderKey)
	fakeBodyHeaderKey              = []byte("fake-body")
	fakeBodyHeaderStringKey        = string(fakeBodyHeaderKey)
)

var (
	SnappyContentEncodingValue       = []byte("snappy")
	SnappyContentEncodingValueString = string(SnappyContentEncodingValue)
	fakeBodyHeaderValue              = []byte("1")
	fakeBodyHeaderValueString        = string(fakeBodyHeaderValue)
)

type HeaderValues struct {
	Name   string
	Values []string
}

var (
	builtinHeaderFieldMap = map[string][]string{
		AgentHeaderStringKey:           nil,
		ForwardedHeaderStringKey:       nil,
		AuthorizationHeaderStringKey:   nil,
		ContentLengthHeaderStringKey:   nil,
		ContentTypeHeaderStringKey:     nil,
		ContentEncodingHeaderStringKey: {SnappyContentEncodingValueString},
		SignatureHeaderStringKey:       nil,
		fakeBodyHeaderStringKey:        {fakeBodyHeaderValueString},
	}
)

func RegisterBuiltinHeaderFields(name string, values ...string) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return
	}
	builtins, has := builtinHeaderFieldMap[name]
	if !has {
		builtinHeaderFieldMap[name] = values
		return
	}
	slices.Sort[[]string](builtins)
	for _, value := range values {
		if i, exist := slices.BinarySearch[[]string](builtins, value); exist {
			values = slices.Delete[[]string](values, i, i+1)
			continue
		}
	}
	if valuesLen := len(values); valuesLen > 0 {
		for _, value := range values {
			builtins = append(builtins, value)
		}
		slices.Sort[[]string](builtins)
	}
}

func builtinHeaderFields() []HeaderValues {
	if len(builtinHeaderFieldMap) == 0 {
		return make([]HeaderValues, 0)
	}
	headers := make([]HeaderValues, 0, len(builtinHeaderFieldMap))
	for name, values := range builtinHeaderFieldMap {
		slices.Sort[[]string](values)
		headers = append(headers, HeaderValues{
			Name:   name,
			Values: values,
		})
	}
	return headers
}

func newHeaderPacker(maxHeaderSize int, headers []HeaderValues) (packer *bpack.Packer, err error) {
	fields := builtinHeaderFields()
	builtinLen := len(fields)

	for _, field := range headers {
		field.Name = strings.ToLower(strings.TrimSpace(field.Name))
		if field.Name == "" {
			continue
		}
		if len(field.Values) > 0 {
			values := make([]string, 0, len(field.Values))
			for i, value := range field.Values {
				value = strings.TrimSpace(value)
				if value == "" {
					continue
				}
				exist := false
				for j := 0; j < i; j++ {
					if strings.TrimSpace(field.Values[j]) == value {
						exist = true
						break
					}
				}
				if exist {
					continue
				}
				values = append(values, value)
			}
			field.Values = values
			slices.Sort[[]string](field.Values)
		}

		var idx int = -1
		for i := 0; i < builtinLen; i++ {
			if fields[i].Name == field.Name {
				idx = i
				break
			}
		}
		if idx == -1 {
			fields = append(fields, field)
			continue
		}
		builtin := fields[idx]
		for _, value := range field.Values {
			n := sort.SearchStrings(builtin.Values, value)
			if n == len(builtin.Values) {
				builtin.Values = append(builtin.Values, value)
				slices.Sort[[]string](builtin.Values)
				fields[idx] = builtin
			}
		}
	}

	options := make([]bpack.Option, 0, len(fields)+1)
	options = append(options, bpack.MaxHeaderSize(maxHeaderSize))
	for _, field := range fields {
		if len(field.Values) > 0 {
			options = append(options, bpack.Field(field.Name, field.Values...))
		} else {
			options = append(options, bpack.Field(field.Name))
		}
	}

	packer, err = bpack.New(options...)
	return
}

type HeaderIterator bpack.HeaderIterator

type Agent struct {
	id     []byte
	device []byte
}

func (agent *Agent) Id() []byte {
	return agent.id
}

func (agent *Agent) Device() []byte {
	return agent.device
}

func (agent *Agent) Encode() []byte {
	if len(agent.id) == 0 && len(agent.device) == 0 {
		return nil
	}
	return bytes.Join([][]byte{agent.id, agent.device}, []byte{';'})
}

func (agent *Agent) Decode(p []byte) error {
	if len(p) == 0 {
		return nil
	}
	i := bytes.IndexByte(p, ';')
	if i == -1 {
		return errors.New("invalid agent header")
	}
	agent.id = p[:i]
	agent.device = p[i+1:]
	return nil
}

type ForwardedEntry struct {
	name  []byte
	host  []byte
	proto []byte
}

func (entry *ForwardedEntry) Encode() []byte {
	return bytes.Join([][]byte{entry.name, entry.host, entry.proto}, []byte{';'})
}

func (entry *ForwardedEntry) Decode(p []byte) error {
	if len(p) == 0 {
		return nil
	}
	items := bytes.Split(p, []byte{';'})
	if len(items) != 3 {
		return errors.New("invalid forwarded entry")
	}
	entry.name = items[0]
	entry.host = items[1]
	entry.proto = items[2]
	return nil
}

type Forwarded struct {
	entries []*ForwardedEntry
}

func (forwarded *Forwarded) Head() (name []byte, host []byte, proto []byte, ok bool) {
	if len(forwarded.entries) == 0 {
		return
	}
	name, host, proto, ok = forwarded.entries[0].name, forwarded.entries[0].host, forwarded.entries[0].proto, true
	return
}

func (forwarded *Forwarded) Add(name []byte, host []byte, proto []byte) {
	forwarded.entries = append(forwarded.entries, &ForwardedEntry{
		name:  name,
		host:  host,
		proto: proto,
	})
}

func (forwarded *Forwarded) Range(f func(name []byte, host []byte, proto []byte) bool) {
	for _, entry := range forwarded.entries {
		if !f(entry.name, entry.host, entry.proto) {
			break
		}
	}
}

func (forwarded *Forwarded) Encode() (p []byte) {
	for i, entry := range forwarded.entries {
		if i > 0 {
			p = append(p, ',')
		}
		p = append(p, entry.Encode()...)
	}
	return
}

func (forwarded *Forwarded) Decode(p []byte) error {
	if len(p) == 0 {
		return nil
	}
	pp := bytes.Split(p, []byte{','})
	for _, b := range pp {
		entry := &ForwardedEntry{}
		if err := entry.Decode(b); err != nil {
			forwarded.entries = forwarded.entries[:0]
			return err
		}
		forwarded.entries = append(forwarded.entries, entry)
	}
	return nil
}

type Header interface {
	Agent() *Agent
	SetAgent(id []byte, device []byte)
	Forwarded() *Forwarded
	AddForwarded(name []byte, host []byte, proto []byte)
	Authorization() []byte
	SetAuthorization(authorization []byte)
	ContentLength() uint64
	SetContentLength(length uint64)
	ContentType() []byte
	SetContentType(typ []byte)
	ContentEncoding() []byte
	SetContentEncoding(encoding []byte)
	Get(key []byte) []byte
	Set(key []byte, value []byte) error
	Remove(key []byte)
	Iterator() HeaderIterator
}

type headerEntry struct {
	key   []byte
	value []byte
}

type header struct {
	noCopy          noCopy
	agent           *Agent
	forwarded       *Forwarded
	authorization   []byte
	contentType     []byte
	contentEncoding []byte
	contentLength   uint64
	entries         []headerEntry
}

func (h *header) Agent() *Agent {
	return h.agent
}

func (h *header) SetAgent(id []byte, device []byte) {
	if h.agent == nil {
		h.agent = new(Agent)
	}
	h.agent.id = id
	h.agent.device = device
}

func (h *header) Forwarded() *Forwarded {
	return h.forwarded
}

func (h *header) AddForwarded(name []byte, host []byte, proto []byte) {
	if h.forwarded == nil {
		h.forwarded = new(Forwarded)
	}
	h.forwarded.Add(name, host, proto)
}

func (h *header) Authorization() []byte {
	return h.authorization
}

func (h *header) SetAuthorization(authorization []byte) {
	h.authorization = authorization
}

func (h *header) ContentLength() uint64 {
	return h.contentLength
}

func (h *header) SetContentLength(length uint64) {
	h.contentLength = length
}

func (h *header) ContentType() []byte {
	return h.contentType
}

func (h *header) SetContentType(typ []byte) {
	h.contentType = typ
}

func (h *header) ContentEncoding() []byte {
	return h.contentEncoding
}

func (h *header) SetContentEncoding(encoding []byte) {
	h.contentEncoding = encoding
}

func (h *header) Get(key []byte) []byte {
	if len(key) == 0 {
		return nil
	}
	sk := unsafe.String(unsafe.SliceData(key), len(key))
	switch sk {
	case AgentHeaderStringKey:
		return h.agent.Encode()
	case ForwardedHeaderStringKey:
		return h.forwarded.Encode()
	case AuthorizationHeaderStringKey:
		return h.authorization
	case ContentLengthHeaderStringKey:
		return quicvarint.Append(nil, h.contentLength)
	case ContentTypeHeaderStringKey:
		return h.contentType
	case ContentEncodingHeaderStringKey:
		return h.contentEncoding
	default:
		for _, kp := range h.entries {
			if unsafe.String(unsafe.SliceData(kp.key), len(kp.key)) == sk {
				return kp.value
			}
		}
		return nil
	}
}

func (h *header) Set(key []byte, value []byte) (err error) {
	if len(key) == 0 {
		return
	}
	sk := unsafe.String(unsafe.SliceData(key), len(key))
	switch sk {
	case AgentHeaderStringKey:
		if h.agent == nil {
			h.agent = new(Agent)
		}
		err = h.agent.Decode(key)
		return
	case ForwardedHeaderStringKey:
		if h.forwarded == nil {
			h.forwarded = new(Forwarded)
		}
		err = h.forwarded.Decode(key)
		return
	case AuthorizationHeaderStringKey:
		h.authorization = value
		return
	case ContentLengthHeaderStringKey:
		h.contentLength, _, err = quicvarint.Parse(value)
		return
	case ContentTypeHeaderStringKey:
		h.contentType = value
		return
	case ContentEncodingHeaderStringKey:
		h.contentEncoding = value
		return
	default:
		for i, kp := range h.entries {
			if unsafe.String(unsafe.SliceData(kp.key), len(kp.key)) == sk {
				h.entries[i].value = value
				return
			}
		}
		h.entries = append(h.entries, headerEntry{key, value})
		return
	}
}

func (h *header) Remove(key []byte) {
	if len(key) == 0 {
		return
	}
	sk := unsafe.String(unsafe.SliceData(key), len(key))
	switch sk {
	case AgentHeaderStringKey:
		h.agent = nil
		return
	case ForwardedHeaderStringKey:
		h.forwarded = nil
		return
	case AuthorizationHeaderStringKey:
		h.authorization = nil
		return
	case ContentLengthHeaderStringKey:
		h.contentLength = 0
		return
	case ContentTypeHeaderStringKey:
		h.contentType = nil
		return
	case ContentEncodingHeaderStringKey:
		h.contentEncoding = nil
		return
	default:
		i := 0
		for ; i < len(h.entries); i++ {
			if unsafe.String(unsafe.SliceData(h.entries[i].key), len(h.entries[i].key)) == sk {
				break
			}
		}
		h.entries = append(h.entries[:i], h.entries[i+1:]...)
		return
	}
}

func (h *header) Iterator() HeaderIterator {
	return func(yield func([]byte, []byte) bool) {
		if h.agent != nil {
			if !yield(AgentHeaderKey, h.agent.Encode()) {
				return
			}
		}
		if h.forwarded != nil {
			if !yield(ForwardedHeaderKey, h.forwarded.Encode()) {
				return
			}
		}
		if len(h.authorization) != 0 {
			if !yield(AuthorizationHeaderKey, h.authorization) {
				return
			}
		}
		if !yield(ContentLengthHeaderKey, quicvarint.Append(nil, h.contentLength)) {
			return
		}
		if len(h.contentType) != 0 {
			if !yield(ContentTypeHeaderKey, h.contentType) {
				return
			}
		}
		if len(h.contentEncoding) != 0 {
			if !yield(ContentEncodingHeaderKey, h.contentEncoding) {
				return
			}
		}
		for _, k := range h.entries {
			if !yield(k.key, k.value) {
				return
			}
		}
	}
}

func (h *header) Reset() {
	if h.agent != nil {
		h.agent.id = nil
		h.agent.device = nil
	}
	if h.forwarded != nil {
		h.forwarded.entries = h.forwarded.entries[:0]
	}
	h.authorization = nil
	h.contentLength = 0
	h.contentType = nil
	h.contentEncoding = nil
	if h.entries != nil {
		h.entries = h.entries[:0]
	}
}

func (h *header) Parse(r io.Reader, packer *bpack.Packer) (err error) {
	if err = packer.UnpackFrom(r, h); err != nil {
		err = errors.Join(ErrReadHeaderFailed, err)
	}
	return
}

func (h *header) Flush(w io.Writer, packer *bpack.Packer) (err error) {
	iter := bpack.HeaderIterator(h.Iterator())
	if err = packer.PackTo(w, iter); err != nil {
		err = errors.Join(ErrWriteHeaderFailed, err)
	}
	return
}

var (
	headerPool = sync.Pool{
		New: func() any {
			return new(header)
		},
	}
)

func AcquireHeader() Header {
	return headerPool.Get().(*header)
}

func ReleaseHeader(h Header) {
	if h == nil {
		return
	}
	if hh, ok := h.(*header); ok {
		hh.Reset()
		headerPool.Put(hh)
	}
}

func UintHeaderValue(u uint64) []byte {
	return quicvarint.Append(nil, u)
}
