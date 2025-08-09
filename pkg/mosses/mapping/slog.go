package mapping

import (
	"context"
	"log/slog"

	"github.com/brickingsoft/brick/pkg/mosses"
)

func SetDefaultSLog(logger mosses.Logger) {
	slog.SetDefault(SLog(logger))
}

func SLog(logger mosses.Logger) (v *slog.Logger) {
	v = slog.New(&SLogHandler{moss: logger.CallerSkipShift(3)})
	return
}

type SLogHandler struct {
	moss mosses.Logger
}

func (handler SLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	switch level {
	case slog.LevelDebug:
		return handler.moss.DebugEnabled()
	case slog.LevelInfo:
		return handler.moss.InfoEnabled()
	case slog.LevelWarn:
		return handler.moss.WarnEnabled()
	case slog.LevelError:
		return handler.moss.ErrorEnabled()
	default:
		return false
	}
}

func (handler SLogHandler) Handle(ctx context.Context, record slog.Record) error {
	switch record.Level {
	case slog.LevelDebug:
		handler.moss.Debug(ctx, record.Message)
		break
	case slog.LevelInfo:
		handler.moss.Info(ctx, record.Message)
		break
	case slog.LevelWarn:
		handler.moss.Warn(ctx, record.Message)
		break
	case slog.LevelError:
		handler.moss.Error(ctx, record.Message)
		break
	default:
		break
	}
	return nil
}

func (handler SLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return handler
	}
	attributes := make([]mosses.Attribute, len(attrs))
	for i, attr := range attrs {
		attributes[i] = mosses.Attribute{
			Key:   attr.Key,
			Value: attr.Value,
		}
	}
	return &SLogHandler{
		moss: handler.moss.Attr(attributes...).CallerSkipShift(3),
	}
}

func (handler SLogHandler) WithGroup(name string) slog.Handler {
	return &SLogHandler{
		moss: handler.moss.Group(name).CallerSkipShift(3),
	}
}
