package whisper

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

type Wind interface {
	os.Signal
}

var (
	SIGHUP  = syscall.SIGHUP
	SIGINT  = syscall.SIGINT
	SIGABRT = syscall.SIGABRT
	SIGKILL = syscall.SIGKILL
	SIGTERM = syscall.SIGTERM
)

func Listen(ctx context.Context, winds ...Wind) (notify context.Context, cancel context.CancelFunc) {
	if len(winds) == 0 {
		winds = []Wind{SIGHUP, SIGINT, SIGABRT, SIGKILL, SIGTERM}
	}
	sigs := make([]os.Signal, 0, 1)
	for _, w := range winds {
		if w == nil {
			continue
		}
		sigs = append(sigs, w)
	}
	if len(sigs) == 0 {
		sigs = append(sigs, SIGHUP, SIGINT, SIGABRT, SIGKILL, SIGTERM)
	}
	notify, cancel = signal.NotifyContext(ctx, sigs...)
	return
}
