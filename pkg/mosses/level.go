package mosses

import (
	"fmt"
	"strings"
)

const (
	debugLevel   = "DEBUG"
	infoLevel    = "INFO"
	warnLevel    = "WARN"
	errorLevel   = "ERROR"
	unknownLevel = "UNKNOWN"
)

const (
	UnknownLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
)

type Level int

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return debugLevel
	case InfoLevel:
		return infoLevel
	case WarnLevel:
		return warnLevel
	case ErrorLevel:
		return errorLevel
	default:
		return unknownLevel
	}
}

func (l Level) Validate() bool {
	return l == DebugLevel || l == InfoLevel || l == WarnLevel || l == ErrorLevel
}

func (l Level) Enabled(target Level) bool {
	if l.Validate() {
		return l <= target
	}
	return false
}

func LevelFromString(s string) (level Level, err error) {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)
	switch s {
	case debugLevel:
		return DebugLevel, nil
	case infoLevel:
		return InfoLevel, nil
	case warnLevel:
		return WarnLevel, nil
	case errorLevel:
		return ErrorLevel, nil
	default:
		return UnknownLevel, fmt.Errorf("unknown level: %s", s)
	}
}
