package mosses

import (
	"context"
	"errors"
	"strings"
)

type Logger interface {
	Group(name string) Logger
	Attr(attrs ...Attribute) Logger
	CallerSkipShift(shift int) Logger
	DebugEnabled() bool
	Debug(ctx context.Context, msg string, args ...any)
	InfoEnabled() bool
	Info(ctx context.Context, msg string, args ...any)
	WarnEnabled() bool
	Warn(ctx context.Context, msg string, args ...any)
	ErrorEnabled() bool
	Error(ctx context.Context, msg string, args ...any)
	Close() error
}

type Options struct {
	Level   Level
	Source  bool
	Group   string
	Handler Handler
}

type Option func(options *Options) (err error)

func WithLevel(level Level) Option {
	return func(options *Options) (err error) {
		if !level.Validate() {
			err = errors.New("mosses: invalid level")
			return
		}
		options.Level = level
		return
	}
}

func WithSource(source bool) Option {
	return func(options *Options) (err error) {
		options.Source = source
		return
	}
}

func WithGroup(group string) Option {
	return func(options *Options) (err error) {
		options.Group = strings.TrimSpace(group)
		return
	}
}

func WithHandler(handler Handler) Option {
	return func(options *Options) (err error) {
		if handler == nil {
			err = errors.New("mosses: missing handler")
			return
		}
		options.Handler = handler
		return
	}
}

func New(options ...Option) (logger Logger, err error) {
	opts := Options{
		Level:   InfoLevel,
		Source:  true,
		Group:   "",
		Handler: nil,
	}
	for _, opt := range options {
		if err = opt(&opts); err != nil {
			err = errors.New("mosses: new failed, " + err.Error())
			return
		}
	}

	if opts.Handler == nil {
		opts.Handler = NewStandardOutColorfulHandler()
	}

	logger = &Moss{
		level:           opts.Level,
		sourced:         opts.Source,
		callerSkipShift: 0,
		group: Group{
			Name:   opts.Group,
			Attrs:  nil,
			Parent: nil,
		},
		handler: opts.Handler,
	}
	return
}
