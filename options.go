package brick

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/brickingsoft/brick/configs"
	"github.com/brickingsoft/brick/endpoints"
	"github.com/brickingsoft/brick/logs"
	"github.com/brickingsoft/brick/pkg/whisper"
	"github.com/brickingsoft/brick/transports"
)

type Options struct {
	Active                      string
	Version                     string
	ConfigRetriever             configs.Retriever
	LoggerBuilder               logs.Builder
	EndpointBuilders            []endpoints.EndpointBuilder
	EndpointRetrieverBuilder    endpoints.EndpointRetrieverBuilder
	ExtraTransportBuilders      []transports.Builder
	GracefulShutdownListenWinds []whisper.Wind
	CloseTimeout                time.Duration
}

type Option func(*Options) error

func WithActive(active string) Option {
	return func(o *Options) error {
		o.Active = strings.TrimSpace(active)
		return nil
	}
}

func WithVersion(v string) Option {
	return func(o *Options) error {
		o.Version = strings.TrimSpace(v)
		return nil
	}
}

func WithConfigRetriever(retriever configs.Retriever) Option {
	return func(o *Options) error {
		o.ConfigRetriever = retriever
		return nil
	}
}

func WithLogger(builder logs.Builder) Option {
	return func(o *Options) error {
		o.LoggerBuilder = builder
		return nil
	}
}

func WithEndpoint(builder ...endpoints.EndpointBuilder) Option {
	return func(o *Options) error {
		o.EndpointBuilders = append(o.EndpointBuilders, builder...)
		return nil
	}
}

func WithEndpoints(builders []endpoints.EndpointBuilder) Option {
	return func(o *Options) error {
		o.EndpointBuilders = builders
		return nil
	}
}

func WithEndpointRetriever(builder endpoints.EndpointRetrieverBuilder) Option {
	return func(o *Options) error {
		o.EndpointRetrieverBuilder = builder
		return nil
	}
}

func WithExtraTransport(builder ...transports.Builder) Option {
	return func(o *Options) error {
		o.ExtraTransportBuilders = builder
		return nil
	}
}

func WithGracefulShutdown(signals ...os.Signal) Option {
	return func(o *Options) error {
		if len(signals) == 0 {
			return nil
		}
		winds := make([]whisper.Wind, 0, 1)
		for i, signal := range signals {
			if signal == nil {
				return fmt.Errorf("graceful shutdown signal %d is nil", i)
			}
			winds = append(winds, signal)
		}
		o.GracefulShutdownListenWinds = winds
		return nil
	}
}

func WithCloseTimeout(timeout time.Duration) Option {
	return func(o *Options) error {
		if timeout < 0 {
			timeout = 0
		}
		o.CloseTimeout = timeout
		return nil
	}
}
