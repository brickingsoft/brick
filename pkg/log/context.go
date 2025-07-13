package log

import (
	"context"
	"fmt"
	"log/slog"
)

type loggerCtxKey struct {
	name string
}

var (
	ctxKey = loggerCtxKey{"$.brick.logger"}
)

func With(ctx context.Context, logger *slog.Logger) context.Context {
	ctx = context.WithValue(ctx, ctxKey, logger)
	return ctx
}

func Logger(ctx context.Context) *slog.Logger {
	v := ctx.Value(ctxKey)
	if v == nil {
		panic(fmt.Sprintf("brick: there is no '%s' in context", ctxKey.name))
	}

	l, ok := v.(*slog.Logger)
	if !ok {
		panic("brick: there is no slog.Logger in context")
	}
	return l
}
