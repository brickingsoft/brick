package contexts_test

import (
	"context"
	"reflect"
	"testing"
	"unsafe"

	"github.com/brickingsoft/brick/pkg/contexts"
)

type Parent struct {
	context.Context
	Name string
}

func (p *Parent) NS() string {
	return p.Name
}

type Context interface {
	context.Context
	NS() string
}

func TestParent(t *testing.T) {
	parent := &Parent{
		Context: context.Background(),
		Name:    "parent",
	}

	ctx := context.WithValue(parent, "1", "1")

	p, ok := contexts.Parent[Context](ctx)
	if ok {
		t.Log("parent:", p.NS())
	} else {
		t.Fatal("failed to find parent")
	}

}

// BenchmarkParent-20    	 9223276	       136.2 ns/op	       0 b/op	       0 allocs/op
func BenchmarkParent(b *testing.B) {
	parent := &Parent{
		Context: context.Background(),
		Name:    "parent",
	}
	ctx := context.WithValue(parent, "1", "1")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok := contexts.Parent[Context](ctx)
		if !ok {
			b.Fatal("failed to find parent")
		}
	}
}

// BenchmarkContext_Value-20    	256423624	         4.532 ns/op	       0 b/op	       0 allocs/op
func BenchmarkContext_Value(b *testing.B) {
	parent := &Parent{
		Context: context.Background(),
		Name:    "parent",
	}
	ctx := context.WithValue(parent, "1", parent)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ctx.Value("1")
	}
}

func TestB(t *testing.T) {
	b := []byte{1, 2, 3}
	p := B(b)
	t.Log(unsafe.SliceData(b) == unsafe.SliceData(p))
	t.Log(reflect.ValueOf(b).UnsafePointer())
	t.Log(reflect.ValueOf(p).UnsafePointer())
}

func B(b []byte) []byte {
	return b
}

func TestS(t *testing.T) {
	s := "123"
	ss := S(s)
	t.Log(reflect.ValueOf(s).UnsafePointer())
	t.Log(reflect.ValueOf(ss).UnsafePointer())
}

func S(s string) string {
	return s
}
