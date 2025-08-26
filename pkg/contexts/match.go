package contexts

import (
	"context"
	"reflect"
)

func Match(ctx context.Context, parent context.Context) bool {
	if parent == nil || ctx == nil {
		return false
	}
	tv := reflect.ValueOf(ctx)
	if tv.Kind() != reflect.Ptr {
		return false
	}
	tv = tv.Elem()
	cv := tv.FieldByName("Context")
	if !cv.IsValid() {
		return false
	}
	pv := reflect.ValueOf(parent)
	if pv.Kind() == reflect.Ptr {
		cve := cv.Elem()
		if cve.Kind() != reflect.Ptr {
			return false
		}
		pvp := pv.UnsafePointer()
		cvp := cve.UnsafePointer()
		if pvp == cvp {
			return true
		}
		c := cv.Interface()
		cc, ok := c.(context.Context)
		if !ok {
			return false
		}
		return Match(cc, parent)
	} else {
		c := cv.Interface()
		if parent == c {
			return true
		}
		cc, ok := c.(context.Context)
		if !ok {
			return false
		}
		return Match(cc, parent)
	}
}
