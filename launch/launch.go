package launch

import (
	"context"
)

type CancelFunc func() error

func Launch(ctx context.Context, options ...Option) (context.Context, CancelFunc, error) {
	return nil, nil, nil
}
