package logs

import (
	"fmt"
	"strings"

	"github.com/brickingsoft/brick/pkg/mosses"
)

type Options struct {
	Level   string         `json:"level" yaml:"level"`
	Source  bool           `json:"source" yaml:"source"`
	Group   string         `json:"group" yaml:"group"`
	Handler HandlerOptions `json:"handler" yaml:"handler"`
}

func New(options Options) mosses.Logger {
	level, levelErr := mosses.LevelFromString(options.Level)
	if levelErr != nil {
		panic(fmt.Sprintf("brick: invalid log level: %v", levelErr))
	}
	source := options.Source
	group := strings.TrimSpace(options.Group)
	handler := newHandler(options.Handler)
	mLogger, err := mosses.New(mosses.WithLevel(level), mosses.WithSource(source), mosses.WithGroup(group), mosses.WithHandler(handler))
	if err != nil {
		panic(fmt.Sprintf("brick: unable to create mosses logger: %v", err))
	}
	return mLogger
}
