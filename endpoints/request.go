package endpoints

import (
	"context"

	"github.com/brickingsoft/brick/transports"
)

type Header interface {
	transports.Header
}

type Request interface {
	Endpoint() string
	Function() string
	Header() Header
	Body() ([]byte, error)
}

type Response interface {
	Succeed() bool
	Header() Header
	Body() (body []byte, err error)
	ParseBody(v any) (err error)
}

type ResponseWriter interface {
	AddHeader(key string, values ...string) ResponseWriter
	GetHeader(key string) (values []string)
	RemoveHeader(key string) ResponseWriter
	Succeed(v any)
	Failed(v error)
}

type responseWriter struct {
	proxy transports.ResponseWriter
}

func (w *responseWriter) AddHeader(key string, values ...string) ResponseWriter {
	if len(key) == 0 || len(values) == 0 {
		return w
	}
	for _, value := range values {
		w.proxy.Header().Add(key, value)
	}
	return w
}

func (w *responseWriter) GetHeader(key string) (values []string) {
	return w.proxy.Header().Values(key)
}

func (w *responseWriter) RemoveHeader(key string) ResponseWriter {
	w.proxy.Header().Remove(key)
	return w
}

func (w *responseWriter) Succeed(v any) {
	w.proxy.Succeed(v)
}

func (w *responseWriter) Failed(v error) {
	w.proxy.Failed(v)
}

type Stream interface {
	Next() (r RequestCtx, ok bool)
	Response() ResponseWriter
	Close() (err error)
}

type stream struct {
	proxy  transports.Stream
	writer ResponseWriter
}

func (s *stream) Next() (r RequestCtx, ok bool) {
	tr, has := s.proxy.Next()
	if !has {
		return
	}
	r = &requestCtx{tr}
	ok = true
	return
}

func (s *stream) Response() ResponseWriter {
	return s.writer
}

func (s *stream) Close() error {
	return s.proxy.Close()
}

type HijackHandler interface {
	Handle(ctx context.Context, stream Stream)
}

func mapToTransportHijackHandler(handler HijackHandler) transports.HijackHandler {
	return &transportHijackHandler{
		proxy: handler,
	}
}

type transportHijackHandler struct {
	proxy HijackHandler
}

func (handler *transportHijackHandler) Handle(ctx context.Context, s transports.Stream) {
	pr := &stream{
		proxy: s,
		writer: &responseWriter{
			proxy: s.Response(),
		},
	}
	handler.proxy.Handle(ctx, pr)
}

type RequestCtx interface {
	context.Context
	Request
	ParseBody(dst any) (err error)
	Response() ResponseWriter
	Hijacked() bool
	Hijack(handler HijackHandler) (err error)
}

type requestCtx struct {
	transports.RequestCtx
}

func (r *requestCtx) Header() Header {
	return r.RequestCtx.Header()
}

func (r *requestCtx) Response() ResponseWriter {
	return r
}

func (r *requestCtx) Hijack(handler HijackHandler) (err error) {
	err = r.RequestCtx.Hijack(mapToTransportHijackHandler(handler))
	return
}

func (r *requestCtx) AddHeader(key string, values ...string) ResponseWriter {
	if len(key) == 0 || len(values) == 0 {
		return r
	}
	response := r.RequestCtx.Response()
	for _, value := range values {
		response.Header().Add(key, value)
	}
	return r
}

func (r *requestCtx) GetHeader(key string) []string {
	return r.RequestCtx.Response().Header().Values(key)
}

func (r *requestCtx) RemoveHeader(key string) ResponseWriter {
	r.RequestCtx.Response().Header().Remove(key)
	return r
}

func (r *requestCtx) Succeed(v any) {
	r.RequestCtx.Response().Succeed(v)
}

func (r *requestCtx) Failed(v error) {
	r.RequestCtx.Response().Failed(v)
}
