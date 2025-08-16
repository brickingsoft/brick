package logs

import (
	"context"
	"errors"

	"github.com/brickingsoft/brick/pkg/mosses"
)

const (
	callerSkipShift = 1
)

var (
	ctxKey = loggerCtxKey{"$.brick.logger"}
)

type loggerCtxKey struct {
	name string
}

func With(ctx context.Context, logger Logger) context.Context {
	ctx = context.WithValue(ctx, ctxKey, logger)
	return ctx
}

func GetLogger(ctx context.Context) Logger {
	v := ctx.Value(ctxKey)
	if v == nil {
		panic(errors.New("context does not contain a logger"))
	}
	logger, ok := v.(Logger)
	if !ok {
		panic(errors.New("context contains a invalid typed logger"))
	}
	return logger
}

func Group(ctx context.Context, name string) context.Context {
	logger := GetLogger(ctx)
	logger = logger.Group(name)
	return With(ctx, logger)
}

func Attr(ctx context.Context, attrs ...Attribute) context.Context {
	if len(attrs) == 0 {
		return ctx
	}
	logger := GetLogger(ctx)
	mas := make([]mosses.Attribute, len(attrs))
	for i, a := range attrs {
		mas[i] = a.Attribute
	}
	logger = logger.Attr(mas...)
	return With(ctx, logger)
}

func CallerSkipShift(ctx context.Context, shift int) context.Context {
	logger := GetLogger(ctx)
	logger = logger.CallerSkipShift(shift)
	return With(ctx, logger)
}

func DebugEnabled(ctx context.Context) bool {
	logger := GetLogger(ctx)
	return logger.DebugEnabled()
}

func Debug(ctx context.Context, msg string, args ...any) {
	logger := GetLogger(ctx)
	logger.CallerSkipShift(callerSkipShift).Debug(ctx, msg, args...)
}

func InfoEnabled(ctx context.Context) bool {
	logger := GetLogger(ctx)
	return logger.InfoEnabled()
}

func Info(ctx context.Context, msg string, args ...any) {
	logger := GetLogger(ctx)
	logger.CallerSkipShift(callerSkipShift).Info(ctx, msg, args...)
}

func WarnEnabled(ctx context.Context) bool {
	logger := GetLogger(ctx)
	return logger.WarnEnabled()
}

func Warn(ctx context.Context, msg string, args ...any) {
	logger := GetLogger(ctx)
	logger.CallerSkipShift(callerSkipShift).Warn(ctx, msg, args...)
}

func ErrorEnabled(ctx context.Context) bool {
	logger := GetLogger(ctx)
	return logger.ErrorEnabled()
}

func Error(ctx context.Context, msg string, args ...any) {
	logger := GetLogger(ctx)
	logger.CallerSkipShift(callerSkipShift).Error(ctx, msg, args...)
}
