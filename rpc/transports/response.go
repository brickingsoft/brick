package transports

import (
	"io"
	"sync"

	"github.com/brickingsoft/brick/pkg/avro"
	"github.com/brickingsoft/brick/rpc/errors"
	"github.com/brickingsoft/bytebuffers"
)

var (
	ErrResponseParseBodyFailed = errors.New("failed to parse Response body")
	ErrResponseWriteBodyFailed = errors.New("failed to write Response body")
)

type Response struct {
	succeed bool
	header  *header
	body    bytebuffers.Buffer
}

func (resp *Response) Succeed() bool {
	return resp.succeed
}

func (resp *Response) Header() Header {
	return resp.header
}

func (resp *Response) ParseBody(v any) (err error) {
	if err = avro.DecodeFrom(resp.body, v); err != nil {
		err = errors.Join(ErrResponseParseBodyFailed, err)
	}
	return
}

func (resp *Response) Reset() {
	resp.succeed = false
	resp.header.Reset()
	resp.body.Reset()
}

var (
	responsePool = sync.Pool{
		New: func() any {
			return &Response{
				succeed: false,
				header:  &header{},
				body:    bytebuffers.NewBuffer(),
			}
		},
	}
)

func AcquireResponse() *Response {
	v := responsePool.Get().(*Response)
	return v
}

func ReleaseResponse(r *Response) {
	if r == nil {
		return
	}
	r.Reset()
	responsePool.Put(r)
}

func parseResponse(r io.Reader, response *Response) (err error) {

	return
}
