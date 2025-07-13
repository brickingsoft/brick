package log

import (
	"context"
	"log/slog"
)

func Group(ctx context.Context, name string) context.Context {
	logger := Logger(ctx)
	logger = logger.WithGroup(name)
	ctx = With(ctx, logger)
	return ctx
}

func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	Logger(ctx).Log(ctx, level, msg, args...)
}

func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	Logger(ctx).LogAttrs(ctx, level, msg, attrs...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	Logger(ctx).DebugContext(ctx, msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	Logger(ctx).InfoContext(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	Logger(ctx).WarnContext(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	Logger(ctx).ErrorContext(ctx, msg, args...)
}
