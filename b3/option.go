package b3

import "log/slog"

type Option func(opts *Options) (err error)

type Options struct {
	LogLevel slog.Level
}
