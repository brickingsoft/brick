package b3

import (
	"context"
)

func Brick(ctx context.Context, options ...Option) (err error) {
	// options
	opts := Options{}
	for _, opt := range options {
		if err = opt(&opts); err != nil {
			return
		}
	}

	return
}
