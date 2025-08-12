package brick

import (
	"context"
	"errors"
)

func Launch(ctx context.Context, options ...Option) (err error) {
	if ctx == nil {
		err = errors.Join(errors.New("brick launch app failed"), errors.New("context is missing"))
		return
	}
	app, appErr := New(options...)
	if appErr != nil {
		err = errors.Join(errors.New("brick launch app failed"), appErr)
		return
	}
	if srvErr := app.Serve(ctx); srvErr != nil {
		err = errors.Join(errors.New("brick launch app failed"), srvErr)
	}
	_ = app.Close()
	return
}
