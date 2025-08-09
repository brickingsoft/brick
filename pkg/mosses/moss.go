package mosses

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

type Moss struct {
	level           Level
	sourced         bool
	callerSkipShift int
	group           Group
	handler         Handler
}

func (moss *Moss) Group(name string) Logger {
	return &Moss{
		level:           moss.level,
		sourced:         moss.sourced,
		callerSkipShift: 0,
		group: Group{
			Name:   name,
			Attrs:  nil,
			Parent: &moss.group,
		},
		handler: moss.handler,
	}
}

func (moss *Moss) Attr(attrs ...Attribute) Logger {
	group := Group{
		Name:   moss.group.Name,
		Attrs:  moss.group.Attrs,
		Parent: moss.group.Parent,
	}
	group.MergeAttributes(attrs)
	return &Moss{
		level:           moss.level,
		sourced:         moss.sourced,
		callerSkipShift: 0,
		group:           group,
		handler:         moss.handler,
	}
}

func (moss *Moss) CallerSkipShift(shift int) Logger {
	return &Moss{
		level:           moss.level,
		sourced:         moss.sourced,
		callerSkipShift: shift,
		group:           moss.group,
		handler:         moss.handler,
	}
}

func (moss *Moss) DebugEnabled() bool {
	return moss.level.Enabled(DebugLevel)
}

func (moss *Moss) Debug(ctx context.Context, msg string, args ...any) {
	moss.log(ctx, DebugLevel, msg, args...)
}

func (moss *Moss) InfoEnabled() bool {
	return moss.level.Enabled(InfoLevel)
}

func (moss *Moss) Info(ctx context.Context, msg string, args ...any) {
	moss.log(ctx, InfoLevel, msg, args...)
}

func (moss *Moss) WarnEnabled() bool {
	return moss.level.Enabled(WarnLevel)
}

func (moss *Moss) Warn(ctx context.Context, msg string, args ...any) {
	moss.log(ctx, WarnLevel, msg, args...)
}

func (moss *Moss) ErrorEnabled() bool {
	return moss.level.Enabled(ErrorLevel)
}

func (moss *Moss) Error(ctx context.Context, msg string, args ...any) {
	moss.log(ctx, ErrorLevel, msg, args...)
}

func (moss *Moss) Close() error {
	return moss.handler.Close()
}

func (moss *Moss) log(ctx context.Context, level Level, msg string, args ...any) {
	if !moss.level.Enabled(level) {
		return
	}
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	var pc uintptr
	if moss.sourced {
		var pcs [1]uintptr
		runtime.Callers(3+moss.callerSkipShift, pcs[:])
		pc = pcs[0]
	}

	record := Record{
		Level:   level,
		Time:    time.Now(),
		Message: msg,
		PC:      pc,
		Group:   moss.group,
	}
	moss.handler.Handle(ctx, &record)
}
