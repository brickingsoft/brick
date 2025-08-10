package b3

import (
	"github.com/brickingsoft/brick/pkg/mosses"
)

type Option func(opts *Options) (err error)

type LoggerOptions struct {
	Level  mosses.Level
	Source bool
}

type Options struct {
	Logger LoggerOptions
}
