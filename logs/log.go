package logs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/brickingsoft/brick/configs"
	"github.com/brickingsoft/brick/pkg/mosses"
)

type Logger interface {
	mosses.Logger
}

type Builder func(config configs.Config) (logger Logger, err error)

type MossHandlerBuilder func(config configs.Config) (handler mosses.Handler, err error)

type MossBuilderOptions struct {
	HandlerBuilder MossHandlerBuilder
}

type MossBuilderOption func(o *MossBuilderOptions) error

func WithMossHandler(builder MossHandlerBuilder) MossBuilderOption {
	return func(o *MossBuilderOptions) error {
		o.HandlerBuilder = builder
		return nil
	}
}

type MossAsyncHandlerConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
	mosses.AsyncHandlerOptions
}

type MossConfig struct {
	Level   string                 `json:"level" yaml:"level"`
	Source  bool                   `json:"source" yaml:"source"`
	Group   string                 `json:"group" yaml:"group"`
	Async   MossAsyncHandlerConfig `json:"async" yaml:"async"`
	Handler configs.Config         `json:"handler" yaml:"handler"`
}

func Moss(options ...MossBuilderOption) (builder Builder) {
	builder = func(config configs.Config) (logger Logger, err error) {
		opts := MossBuilderOptions{}
		for _, opt := range options {
			if err = opt(&opts); err != nil {
				err = errors.Join(errors.New("build moss logger failed"), err)
				return
			}
		}
		handlerBuilder := opts.HandlerBuilder
		if handlerBuilder == nil {
			handlerBuilder = mossStdoutHandlerBuilder
		}
		mossConfig := MossConfig{}
		if err = config.Node("logger").As(&mossConfig); err != nil {
			err = errors.Join(errors.New("build moss logger failed"), err)
			return
		}
		handler, handlerErr := handlerBuilder(mossConfig.Handler)
		if handlerErr != nil {
			err = errors.Join(errors.New("build moss logger failed"), handlerErr)
			return
		}
		if mossConfig.Async.Enabled {
			handler = mosses.NewAsyncHandler(handler, mossConfig.Async.AsyncHandlerOptions)
		}
		if mossConfig.Level == "" {
			mossConfig.Level = mosses.InfoLevel.String()
		}
		level, levelErr := mosses.LevelFromString(mossConfig.Level)
		if levelErr != nil {
			err = errors.Join(errors.New("build moss logger failed"), levelErr)
			return
		}
		source := mossConfig.Source
		group := strings.TrimSpace(mossConfig.Group)
		logger, err = mosses.New(mosses.WithLevel(level), mosses.WithSource(source), mosses.WithGroup(group), mosses.WithHandler(handler))
		if err != nil {
			err = errors.Join(errors.New("build moss logger failed"), err)
			return
		}
		return
	}
	return
}

type MossHandlerConfig struct {
	Encoder string `json:"encoder" yaml:"encoder"`
}

func mossStdoutHandlerBuilder(config configs.Config) (handler mosses.Handler, err error) {
	options := MossHandlerConfig{}
	if err = config.As(&options); err != nil {
		return
	}
	encoder := strings.TrimSpace(options.Encoder)
	if encoder == "" {
		encoder = "colorful"
	}
	switch encoder {
	case "text":
		handler = mosses.NewStandardOutHandler()
		break
	case "colorful":
		handler = mosses.NewStandardOutColorfulHandler()
		break
	case "json":
		handler = mosses.NewStandardOutJsonHandler()
		break
	default:
		err = fmt.Errorf("unsupported encoder %s", encoder)
		break
	}
	return
}
