package mosses

import "time"

type Attribute struct {
	Key   string
	Value any
}

func String(key string, value any) Attribute {
	return Attribute{Key: key, Value: value}
}

func Bool(key string, value bool) Attribute {
	return Attribute{Key: key, Value: value}
}

func Int(key string, value int) Attribute {
	return Int64(key, int64(value))
}

func Int32(key string, value int32) Attribute {
	return Int64(key, int64(value))
}

func Int64(key string, value int64) Attribute {
	return Attribute{Key: key, Value: value}
}

func Uint(key string, value uint) Attribute {
	return Uint64(key, uint64(value))
}

func Uint32(key string, value uint32) Attribute {
	return Uint64(key, uint64(value))
}

func Uint64(key string, value uint64) Attribute {
	return Attribute{Key: key, Value: value}
}

func Float32(key string, value float32) Attribute {
	return Float64(key, float64(value))
}

func Float64(key string, value float64) Attribute {
	return Attribute{Key: key, Value: value}
}

func Time(key string, value time.Time) Attribute {
	return Attribute{Key: key, Value: value}
}

func Duration(key string, value time.Duration) Attribute {
	return Attribute{Key: key, Value: value}
}

const (
	errAttrKey = "error"
)

func Error(err error) Attribute {
	return Attribute{Key: errAttrKey, Value: err}
}

type Group struct {
	Name   string
	Attrs  []Attribute
	Parent *Group
}

func (group *Group) MergeAttributes(attrs []Attribute) {
	if len(attrs) == 0 {
		return
	}
	if len(group.Attrs) == 0 {
		group.Attrs = append(group.Attrs, attrs...)
		return
	}
	merged := false
	for _, attr := range attrs {
		for _, exist := range group.Attrs {
			if exist.Key == attr.Key {
				exist.Value = attr.Value
				merged = true
				break
			}
		}
		if merged {
			continue
		}
		merged = false
		group.Attrs = append(group.Attrs, attr)
	}
}
