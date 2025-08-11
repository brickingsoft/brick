package transports

import (
	"context"
	"errors"
)

var (
	ParseBodyFailed = errors.New("failed to parse body")
	WriteBodyFailed = errors.New("failed to write body")
)

type Header interface {
	Get(key string) (value string)
	Keys() (keys []string)
	Values(key string) (values []string)
	Set(key string, value string)
	Add(key string, values ...string)
	Remove(key string)
	Authorization() string
}

type Request interface {
	Endpoint() string
	Function() string
	Header() Header
	Body() (body []byte, err error)
}

type Response interface {
	Succeed() bool
	Header() Header
	Body() (body []byte, err error)
	ParseBody(v any) (err error)
}

type ResponseWriter interface {
	Header() Header
	Succeed(v any)
	Failed(err error)
}

type Stream interface {
	Context() context.Context
	Next() (r RequestCtx, ok bool)
	Response() ResponseWriter
	Close() (err error)
}

type HijackHandler interface {
	Handle(ctx context.Context, stream Stream)
}

type RequestCtx interface {
	context.Context
	Request
	ParseBody(v any) (err error)
	Response() (response ResponseWriter)
	Hijacked() bool
	Hijack(handler HijackHandler) (err error)
}
