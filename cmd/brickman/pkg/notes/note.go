package notes

import (
	"context"
	"strings"
	"sync"

	"github.com/brickingsoft/brick/pkg/mosses"
	"github.com/urfave/cli/v3"
)

const (
	callerSkipShift = 1
)

var (
	logger mosses.Logger
	once   sync.Once
)

func New(c *cli.Command) mosses.Logger {
	once.Do(func() {
		level := strings.ToUpper(strings.TrimSpace(c.String("log")))
		if level == "" {
			level = "ERROR"
		}
		lvl, lvlErr := mosses.LevelFromString(level)
		if lvlErr != nil {
			lvl = mosses.ErrorLevel
		}
		logger, _ = mosses.New(mosses.WithLevel(lvl))
	})
	return logger
}

func Logger() mosses.Logger {
	return logger
}

func DebugEnabled(ctx context.Context) bool {
	return logger.DebugEnabled()
}

func Debug(ctx context.Context, msg string, args ...any) {
	logger.CallerSkipShift(callerSkipShift).Debug(ctx, msg, args...)
}

func InfoEnabled(ctx context.Context) bool {
	return logger.InfoEnabled()
}

func Info(ctx context.Context, msg string, args ...any) {
	logger.CallerSkipShift(callerSkipShift).Info(ctx, msg, args...)
}

func WarnEnabled(ctx context.Context) bool {
	return logger.WarnEnabled()
}

func Warn(ctx context.Context, msg string, args ...any) {
	logger.CallerSkipShift(callerSkipShift).Warn(ctx, msg, args...)
}

func ErrorEnabled(ctx context.Context) bool {
	return logger.ErrorEnabled()
}

func Error(ctx context.Context, msg string, args ...any) {
	logger.CallerSkipShift(callerSkipShift).Error(ctx, msg, args...)
}
