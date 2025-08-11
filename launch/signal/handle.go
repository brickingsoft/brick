package signal

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Options struct {
	Signals []os.Signal
	Timeout time.Duration
}

type Option func(*Options) error

func WithSignals(signals ...os.Signal) Option {
	return func(o *Options) error {
		o.Signals = signals
		return nil
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) error {
		if timeout < 1 {
			return errors.New("timeout must be greater than zero")
		}
		o.Timeout = timeout
		return nil
	}
}

var (
	defaultSignals = []os.Signal{
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGABRT,
		syscall.SIGQUIT,
		syscall.SIGABRT,
		syscall.SIGTERM,
	}
)

type Handler interface {
	Handle(ctx context.Context) error
	Close() error
}

func Execute(ctx context.Context, handler Handler, options ...Option) (err error) {
	if ctx == nil {
		err = errors.Join(errors.New("signal execute failed"), errors.New("ctx is nil"))
		return
	}

	if handler == nil {
		err = errors.Join(errors.New("signal execute failed"), errors.New("handler is nil"))
		return
	}

	opts := Options{}
	for _, o := range options {
		if err = o(&opts); err != nil {
			err = errors.Join(errors.New("signal execute failed"), err)
			return
		}
	}
	signals := opts.Signals
	if len(signals) == 0 {
		signals = defaultSignals
	}

	var stop context.CancelFunc
	ctx, stop = signal.NotifyContext(ctx, signals...)
	err = execute(ctx, stop, handler, opts.Timeout)
	return
}

type CancelFunc func() error

func AsyncExecute(ctx context.Context, handler Handler, options ...Option) (stop CancelFunc, err error) {
	if ctx == nil {
		err = errors.Join(errors.New("signal async execute failed"), errors.New("ctx is nil"))
		return
	}

	if handler == nil {
		err = errors.Join(errors.New("signal async execute failed"), errors.New("handler is nil"))
		return
	}

	opts := Options{}
	for _, o := range options {
		if err = o(&opts); err != nil {
			err = errors.Join(errors.New("signal async execute failed"), err)
			return
		}
	}
	signals := opts.Signals
	if len(signals) == 0 {
		signals = defaultSignals
	}
	var cancel context.CancelFunc
	ctx, cancel = signal.NotifyContext(ctx, signals...)
	failed := make(chan error, 1)
	stop = func() error {
		cancel()
		return <-failed
	}
	go func(ctx context.Context, cancel context.CancelFunc, handler Handler, timeout time.Duration, failed chan<- error) {
		execErr := execute(ctx, cancel, handler, timeout)
		if execErr != nil {
			failed <- execErr
		}
		close(failed)
	}(ctx, cancel, handler, opts.Timeout, failed)

	return
}

func execute(ctx context.Context, cancel context.CancelFunc, handler Handler, timeout time.Duration) (err error) {
	defer cancel()
	failed := make(chan error, 1)
	go func(ctx context.Context, handler Handler, failed chan error) {
		if handleErr := handler.Handle(ctx); handleErr != nil {
			failed <- handleErr
		}
	}(ctx, handler, failed)

	select {
	case <-ctx.Done():
		if timeout > 0 {
			timer := time.NewTimer(timeout)
			done := make(chan error, 1)
			go func(handler Handler, done chan<- error) {
				if closeErr := handler.Close(); closeErr != nil {
					done <- closeErr
				}
				close(done)
			}(handler, done)
			select {
			case <-timer.C:
				err = errors.New("close timed out")
				break
			case err = <-done:
				break
			}
			timer.Stop()
		} else {
			err = handler.Close()
		}
		break
	case err = <-failed:
		break
	}
	return
}
