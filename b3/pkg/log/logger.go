package log

import (
	"context"
	"log/slog"
	"os"
)

type loggerCtxKey struct {
	name string
}

var (
	ctxKey = loggerCtxKey{"b3"}
)

func New(ctx context.Context, level slog.Level) context.Context {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}).WithGroup("b3")

	logger := slog.New(handler)

	ctx = context.WithValue(ctx, ctxKey, logger)
	return ctx
}

func Logger(ctx context.Context) *slog.Logger {
	v := ctx.Value(ctxKey)
	if v == nil {
		panic("b3: there is no log in context")
	}

	l, ok := v.(*slog.Logger)
	if !ok {
		panic("b3: there is no slog.Logger in context")
	}
	return l
}
