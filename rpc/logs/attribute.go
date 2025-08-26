package logs

import (
	"time"

	"github.com/brickingsoft/brick/pkg/mosses"
)

type Attribute struct {
	mosses.Attribute
}

func String(key string, value any) Attribute {
	attr := mosses.String(key, value)
	return Attribute{attr}
}

func Bool(key string, value bool) Attribute {
	attr := mosses.Bool(key, value)
	return Attribute{attr}
}

func Int(key string, value int) Attribute {
	attr := mosses.Int(key, value)
	return Attribute{attr}
}

func Int32(key string, value int32) Attribute {
	attr := mosses.Int32(key, value)
	return Attribute{attr}
}

func Int64(key string, value int64) Attribute {
	attr := mosses.Int64(key, value)
	return Attribute{attr}
}

func Uint(key string, value uint) Attribute {
	attr := mosses.Uint(key, value)
	return Attribute{attr}
}

func Uint32(key string, value uint32) Attribute {
	attr := mosses.Uint32(key, value)
	return Attribute{attr}
}

func Uint64(key string, value uint64) Attribute {
	attr := mosses.Uint64(key, value)
	return Attribute{attr}
}

func Float32(key string, value float32) Attribute {
	attr := mosses.Float32(key, value)
	return Attribute{attr}
}

func Float64(key string, value float64) Attribute {
	attr := mosses.Float64(key, value)
	return Attribute{attr}
}

func Time(key string, value time.Time) Attribute {
	attr := mosses.Time(key, value)
	return Attribute{attr}
}

func Duration(key string, value time.Duration) Attribute {
	attr := mosses.Duration(key, value)
	return Attribute{attr}
}

func Err(err error) Attribute {
	attr := mosses.Err(err)
	return Attribute{attr}
}

func Any(value any) Attribute {
	attr := mosses.Any(value)
	return Attribute{attr}
}
