package signal

import (
	"context"
	"errors"
	"os/signal"
	"syscall"
	"time"
)

type Server interface {
	Serve(ctx context.Context) error
}

type CloseableServer interface {
	Server
	Close() error
}

func Wrap(srv CloseableServer) Server {
	return &server{
		srv:     srv,
		timeout: 0,
	}
}

func WrapWithTimeout(srv CloseableServer, timeout time.Duration) Server {
	return &server{
		srv:     srv,
		timeout: timeout,
	}
}

type server struct {
	srv     CloseableServer
	timeout time.Duration
}

func (srv *server) Serve(ctx context.Context) error {
	return ListenWithTimeout(ctx, srv.srv, srv.timeout)
}

func Listen(ctx context.Context, srv CloseableServer) error {
	return ListenWithTimeout(ctx, srv, 0)
}

func ListenWithTimeout(ctx context.Context, srv CloseableServer, timeout time.Duration) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if srv == nil {
		err = errors.New("listen failed for server is nil")
	}

	done := make(chan struct{}, 1)
	errCh := make(chan error, 1)

	go func(ctx context.Context, srv CloseableServer, timeout time.Duration, done chan struct{}, errCh chan error) {

		var stop context.CancelFunc
		ctx, stop = signal.NotifyContext(
			ctx,
			syscall.SIGINT,
			syscall.SIGKILL,
			syscall.SIGABRT,
			syscall.SIGQUIT,
			syscall.SIGABRT,
			syscall.SIGTERM)
		go func(ctx context.Context, cancel context.CancelFunc, srv CloseableServer, errCh chan error) {
			srvErr := srv.Serve(ctx)
			if srvErr != nil {
				errCh <- srvErr
				stop()
			}
		}(ctx, stop, srv, errCh)

		<-ctx.Done()
		if len(errCh) > 0 {
			return
		}
		stop()

		if timeout > 0 {
			timeoutErrCh := make(chan error, 1)
			timer := time.NewTimer(timeout)
			defer timer.Stop()
			go func(srv CloseableServer, timer *time.Timer, timeoutErrCh chan error) {
				if closeErr := srv.Close(); closeErr != nil {
					timeoutErrCh <- closeErr
				}
				close(timeoutErrCh)
			}(srv, timer, timeoutErrCh)
			select {
			case <-timer.C:
				errCh <- errors.New("close timeout")
				break
			case timeoutCloseErr := <-timeoutErrCh:
				if timeoutCloseErr != nil {
					errCh <- timeoutCloseErr
				} else {
					done <- struct{}{}
				}
				break
			}
			return
		}

		if closeErr := srv.Close(); closeErr != nil {
			errCh <- closeErr
		} else {
			done <- struct{}{}
		}
	}(ctx, srv, timeout, done, errCh)

	select {
	case <-done:
		return
	case err = <-errCh:
		return
	}
}
