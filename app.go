package brick

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/brickingsoft/brick/configs"
	"github.com/brickingsoft/brick/endpoints"
	"github.com/brickingsoft/brick/logs"
	"github.com/brickingsoft/brick/pkg/whisper"
	"github.com/brickingsoft/brick/transports"
)

type Config struct {
	Logger     configs.Config `yaml:"logger"`
	Transports configs.Config `yaml:"transports"`
	Endpoints  configs.Config `yaml:"endpoints"`
}

func New(options ...Option) (app *App, err error) {
	opts := &Options{}
	for _, option := range options {
		if err = option(opts); err != nil {
			err = errors.Join(errors.New("new app failed"), err)
			return
		}
	}
	// todo build
	// config

	// logger

	// endpoints
	eps := make([]endpoints.Endpoint, 0, 1)

	// transport
	trs := make([]transports.Transport, 0, 1)

	// discovery

	app = &App{
		locker:   new(sync.Mutex),
		launched: true,
		eps:      eps,
		trs:      trs,
	}

	return
}

type App struct {
	locker       sync.Locker
	launched     bool
	logger       logs.Logger
	eps          []endpoints.Endpoint
	trs          []transports.Transport
	winds        []whisper.Wind
	closeTimeout time.Duration
}

func (app *App) prepare(ctx context.Context) context.Context {
	// todo
	return ctx
}

func (app *App) Serve(ctx context.Context) (err error) {
	if app.launched {
		err = errors.Join(errors.New("app run failed"), errors.New("app already launched"))
		return
	}
	if ctx == nil {
		err = errors.Join(errors.New("app serve failed"), errors.New("context is missing"))
		return
	}
	var cancel context.CancelFunc
	ctx, cancel = whisper.Listen(ctx, app.winds...)
	defer cancel()

	// todo prepare
	// todo transport serve...

	<-ctx.Done()
	return
}

func (app *App) Run(ctx context.Context, handler func(ctx context.Context)) (err error) {
	if ctx == nil {
		err = errors.Join(errors.New("app run failed"), errors.New("context is missing"))
		return
	}
	if handler == nil {
		err = errors.Join(errors.New("app run failed"), errors.New("handler is missing"))
		return
	}

	var cancel context.CancelFunc
	ctx, cancel = whisper.Listen(ctx, app.winds...)
	defer cancel()

	ctx = app.prepare(ctx)
	handler(ctx)
	return
}

func (app *App) Close() (err error) {
	app.locker.Lock()
	defer app.locker.Unlock()
	if !app.launched {
		return
	}
	ctx := context.TODO()
	if timeout := app.closeTimeout; timeout > 0 {
		failed := make(chan error, 1)
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		go func(ctx context.Context, failed chan<- error, app *App) {
			if exitErr := app.exit(ctx); exitErr != nil {
				failed <- exitErr
			}
			close(failed)
		}(ctx, failed, app)
		select {
		case <-ctx.Done():
			err = errors.New("timeout")
			break
		case err = <-failed:
			break
		}
		cancel()
		if err != nil {
			err = errors.Join(errors.New("close app failed"), err)
		}
	}
	return
}

func (app *App) exit(ctx context.Context) (err error) {
	errs := make([]error, 0, 1)
	// prepare
	ctx = app.prepare(ctx)
	// discovery
	// transports
	// endpoints
	// logger

	// err
	if len(errs) > 0 {
		err = errors.Join(errs...)
	}
	return
}
