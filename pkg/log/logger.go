package log

import (
	"log/slog"
	"os"
)

type Options struct {
	Group    string
	Level    slog.Level
	Sourcing bool
	Handler  slog.Handler
}

type Option func(opts *Options) (err error)

func New(options ...Option) (v *slog.Logger, err error) {
	// todo options
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})
	v = slog.New(handler)
	return
}
