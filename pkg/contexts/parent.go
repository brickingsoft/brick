package contexts

import (
	"context"
	"reflect"
)

func Parent[P context.Context](ctx context.Context) (p P, ok bool) {
	if ctx == nil {
		return
	}
	rv := reflect.ValueOf(ctx)
	if rv.Kind() != reflect.Ptr {
		return
	}
	rv = rv.Elem()
	cv := rv.FieldByName("Context")
	if !cv.IsValid() {
		return
	}
	if !cv.CanInterface() {
		return
	}
	c := cv.Interface().(context.Context)
	ct := reflect.TypeOf(c)
	pt := reflect.TypeFor[P]()
	if ct.ConvertibleTo(pt) {
		p = c.(P)
		ok = true
		return
	}
	p, ok = Parent[P](c)
	return
}
