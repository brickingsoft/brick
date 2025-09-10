package transports

import (
	"crypto/rand"
	"io"
	"sync"

	"github.com/brickingsoft/brick/pkg/avro"
	"github.com/brickingsoft/brick/rpc/errors"
	"github.com/brickingsoft/bytebuffers"
)

var (
	ErrRequestParseBodyFailed = errors.New("failed to parse request body")
	ErrRequestSetBodyFailed   = errors.New("failed to set request body")
)

type Request struct {
	function uint64
	header   *header
	body     bytebuffers.Buffer
}

func (r *Request) Function() uint64 {
	return r.function
}

func (r *Request) Header() Header {
	return r.header
}

func (r *Request) ParseBody(v any) (err error) {
	if err = avro.DecodeFrom(r.body, v); err != nil {
		err = errors.Join(ErrRequestParseBodyFailed, err)
		return
	}
	return
}

func (r *Request) SetBody(v any) (err error) {
	if err = avro.EncodeTo(r.body, v); err != nil {
		err = errors.Join(ErrRequestSetBodyFailed, err)
		return
	}
	return
}

func (r *Request) Body() io.Reader {
	return r.body
}

func (r *Request) Sign(signer func(b []byte) (signature []byte)) {
	bLen := r.body.Len()
	if bLen == 0 {
		fake, _ := r.body.Borrow(4)
		_, _ = rand.Read(fake)
		r.body.Return(4)
		signature := signer(fake)
		_ = r.header.Set(SignatureHeaderKey, signature)
		_ = r.header.Set(fakeBodyHeaderKey, fakeBodyHeaderValue)
		return
	}
	p := r.body.Peek(bLen)
	signature := signer(p)
	_ = r.header.Set(SignatureHeaderKey, signature)
	return
}

func (r *Request) Reset() {
	r.function = 0
	r.header.Reset()
	r.body.Reset()
}

var (
	requestPool = sync.Pool{
		New: func() any {
			return &Request{
				function: 0,
				header:   &header{},
				body:     bytebuffers.NewBuffer(),
			}
		},
	}
)

func AcquireRequest(function uint64) *Request {
	v := requestPool.Get().(*Request)
	v.function = function
	return v
}

func ReleaseRequest(r *Request) {
	if r == nil {
		return
	}
	r.Reset()
	requestPool.Put(r)
}

func writeRequest(w io.Writer, request *Request) (err error) {

	return
}
