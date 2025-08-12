package transports

import (
	"context"
	"net"

	"github.com/brickingsoft/brick/configs"
)

type ListenerBuilder func(ctx context.Context, config configs.Config) (ln net.Listener, err error)

var (
	// listenerBuilders
	// such as iouring, gm_tls_listener
	listenerBuilders = make(map[string]ListenerBuilder)
)

func RegisterListenerBuilder(name string, builder ListenerBuilder) {
	listenerBuilders[name] = builder
}

func RetrieveListenerBuilder(name string) ListenerBuilder {
	ln, _ := listenerBuilders[name]
	return ln
}
