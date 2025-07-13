package b3

import (
	"context"

	"github.com/brickingsoft/brick/b3/pkg/log"
)

func Brick(ctx context.Context, options ...Option) (err error) {
	// options
	opts := Options{}
	for _, opt := range options {
		if err = opt(&opts); err != nil {
			return
		}
	}
	// logger
	ctx = log.New(ctx, opts.LogLevel)
	log.Debug(ctx, "...")

	return
}
