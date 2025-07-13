package log

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type loggerCtxKey struct {
	name string
}

var (
	ctxKey = loggerCtxKey{"brick"}
)

func New(ctx context.Context, level string) context.Context {

	var sLevel slog.Level
	level = strings.ToLower(strings.TrimSpace(level))
	switch level {
	case "debug":
		sLevel = slog.LevelDebug
	case "info":
		sLevel = slog.LevelInfo
	case "warn":
		sLevel = slog.LevelWarn
	default:
		sLevel = slog.LevelError
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: sLevel,
	}).WithGroup("brick")

	logger := slog.New(handler)

	ctx = context.WithValue(ctx, ctxKey, logger)
	return ctx
}

func Logger(ctx context.Context) *slog.Logger {
	v := ctx.Value(ctxKey)
	if v == nil {
		panic("brick: there is no log in context")
	}

	l, ok := v.(*slog.Logger)
	if !ok {
		panic("brick: there is no slog.Logger in context")
	}
	return l
}
