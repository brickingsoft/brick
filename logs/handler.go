package logs

import (
	"fmt"
	"strings"

	"github.com/brickingsoft/brick/pkg/mists"
	"github.com/brickingsoft/brick/pkg/mosses"
)

type HandlerBuilder func(config mists.Mist) (handler mosses.Handler, err error)

var (
	handlerBuilders = make(map[string]HandlerBuilder)
)

func RegisterHandlerBuilder(name string, builder HandlerBuilder) {
	handlerBuilders[name] = builder
}

func RetrieveHandlerBuilder(name string) (HandlerBuilder, bool) {
	v, has := handlerBuilders[name]
	return v, has
}

type AsyncHandlerOptions struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
	mosses.AsyncHandlerOptions
}

type StandardOutHandlerOptions struct {
	Encoder string `json:"encoder" yaml:"encoder"`
}

type HandlerOptions struct {
	Name    string              `json:"name" yaml:"name"`
	Async   AsyncHandlerOptions `json:"async" yaml:"async"`
	Options mists.Raw           `json:"options" yaml:"options"`
}

func newHandler(options HandlerOptions) mosses.Handler {

	name := strings.TrimSpace(options.Name)
	if name == "" {
		name = "stdout"
	}
	var handler mosses.Handler
	switch name {
	case "stdout":
		stdoutOptions := StandardOutHandlerOptions{}
		stdoutConfig, stdoutConfigErr := mists.New(options.Options)
		if stdoutConfigErr != nil {
			panic(fmt.Sprintf("brick: unable to read log handler options for %s, %v", name, stdoutConfigErr))
		}
		if decodeErr := stdoutConfig.Decode(&options); decodeErr != nil {
			panic(fmt.Sprintf("brick: unable to decode log handler options for %s, %v", name, decodeErr))
		}
		encoder := strings.TrimSpace(stdoutOptions.Encoder)
		if encoder == "" {
			encoder = "colorful"
		}
		switch encoder {
		case "text":
			handler = mosses.NewStandardOutHandler()
		case "colorful":
			handler = mosses.NewStandardOutColorfulHandler()
		case "json":
			handler = mosses.NewStandardOutJsonHandler()
			break
		default:
			panic(fmt.Sprintf("brick: unknown stdout log record encoder %s", encoder))
		}
		break
	default:
		builder := handlerBuilders[name]
		if builder == nil {
			panic(fmt.Sprintf("brick: no log handler builder registered for %s", name))
		}
		handlerConfig, handlerConfigErr := mists.New(options.Options)
		if handlerConfigErr != nil {
			panic(fmt.Sprintf("brick: unable to load handler configuration for %s, %v", name, handlerConfigErr))
		}
		var buildErr error
		handler, buildErr = builder(handlerConfig)
		if buildErr != nil {
			panic(fmt.Sprintf("brick: unable to build log handler by registered for %s, %v", name, buildErr))
		}
		break
	}
	if options.Async.Enabled {
		handler = mosses.NewAsyncHandler(handler, options.Async.AsyncHandlerOptions)
	}
	return handler
}
