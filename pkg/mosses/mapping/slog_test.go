package mapping_test

import (
	"log/slog"
	"testing"

	"github.com/brickingsoft/brick/pkg/mosses"
	"github.com/brickingsoft/brick/pkg/mosses/mapping"
)

func TestSLog(t *testing.T) {
	moss, _ := mosses.New(mosses.WithLevel(mosses.DebugLevel))
	logger := mapping.SLog(moss)
	logger.WithGroup("group").Info("hello world")
	logger.WithGroup("group").With("key", "value").Info("hello world")
	logger.WithGroup("group").With("key", "value").WithGroup("haha").With("s", "a").Info("hello world")
}

func TestSetDefaultSLog(t *testing.T) {
	moss, _ := mosses.New(mosses.WithLevel(mosses.DebugLevel))
	mapping.SetDefaultSLog(moss)
	slog.Info("hello world")
}
