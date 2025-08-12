package brick

import (
	"context"
	"errors"
)

func Run(ctx context.Context, handler func(ctx context.Context), options ...Option) (err error) {
	if ctx == nil {
		err = errors.Join(errors.New("brick run app failed"), errors.New("context is missing"))
		return
	}
	if handler == nil {
		err = errors.Join(errors.New("brick run app failed"), errors.New("handler is missing"))
		return
	}
	app, appErr := New(options...)
	if appErr != nil {
		err = errors.Join(errors.New("brick run app failed"), appErr)
		return
	}
	if runErr := app.Run(ctx, handler); runErr != nil {
		err = errors.Join(errors.New("brick run app failed"), runErr)
	}
	_ = app.Close()
	return
}
