package endpoints

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/brickingsoft/brick/configs"
	"github.com/brickingsoft/brick/transports"
)

type EndpointInitializer interface {
	Init(ctx context.Context) (err error)
}

type Endpoint interface {
	Name() string
	Handle(ctx RequestCtx)
	Close() (err error)
}

type EndpointBuilder func(ctx context.Context, config configs.Config) (endpoint Endpoint, err error)

type EndpointRetriever interface {
	Retrieve(ctx context.Context, name string) (endpoint Endpoint)
}

type EndpointRetrieverBuilder func(ctx context.Context, entries []Endpoint, config configs.Config) (retriever EndpointRetriever, err error)

func DefaultEndpointRetrieverBuilder(_ context.Context, entries []Endpoint, _ configs.Config) (retriever EndpointRetriever, err error) {
	kvs := make(map[string]Endpoint)
	for _, entry := range entries {
		kvs[entry.Name()] = entry
	}
	retriever = &DefaultEndpointRetriever{entries: kvs}
	return
}

type DefaultEndpointRetriever struct {
	entries map[string]Endpoint
}

func (r *DefaultEndpointRetriever) Retrieve(_ context.Context, name string) (endpoint Endpoint) {
	endpoint = r.entries[name]
	return
}

type Options struct {
	Builder EndpointRetrieverBuilder
}

func New(ctx context.Context, entries []Endpoint, options Options) (eps *Endpoints, err error) {
	builder := options.Builder
	if builder == nil {
		builder = DefaultEndpointRetrieverBuilder
	}
	retriever, retrieverErr := builder(ctx, entries, configs.Config{})
	if retrieverErr != nil {
		err = errors.Join(errors.New("failed to build endpoints"), retrieverErr)
		return
	}
	eps = &Endpoints{
		entries:   entries,
		retriever: retriever,
	}
	return
}

type Endpoints struct {
	entries   []Endpoint
	retriever EndpointRetriever
	requests  sync.Pool
}

func (e *Endpoints) acquireRequest(ctx transports.RequestCtx) *requestCtx {
	v := e.requests.Get()
	if v == nil {
		return &requestCtx{ctx}
	}
	req := v.(*requestCtx)
	req.RequestCtx = ctx
	return req
}

func (e *Endpoints) releaseRequest(ctx *requestCtx) {
	if ctx.Hijacked() {
		return
	}
	ctx.RequestCtx = nil
	e.requests.Put(ctx)
}

func (e *Endpoints) Handle(ctx transports.RequestCtx) {
	name := ctx.Endpoint()
	ep := e.retriever.Retrieve(ctx, name)
	if ep == nil {
		ctx.Response().Failed(fmt.Errorf("endpoint %s not found", name))
		return
	}
	r := e.acquireRequest(ctx)
	ep.Handle(r)
	e.releaseRequest(r)
}

func (e *Endpoints) Close() (err error) {
	var errs []error
	for _, entry := range e.entries {
		if closeErr := entry.Close(); closeErr != nil {
			if len(errs) == 0 {
				errs = append(errs, errors.New("failed to close endpoints"))
			}
			errs = append(errs, closeErr)
		}
	}
	if len(errs) > 0 {
		err = errors.Join(errs...)
	}
	return
}
