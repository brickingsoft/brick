package transports

import (
	"context"

	"github.com/brickingsoft/brick/configs"
)

type ServeHandler interface {
	Handle(r RequestCtx)
}

type Client interface {
	Do(ctx context.Context, request Request) (res Response, err error)
	Close() (err error)
}

type Transport interface {
	Name() string
	Listen(ctx context.Context, handler ServeHandler) (err error)
	Connect(ctx context.Context, address string) (client Client, err error)
	Close() (err error)
}

type Builder func(ctx context.Context, config configs.Config) (transport Transport, err error)
