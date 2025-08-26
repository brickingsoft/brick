package bytebuffer_test

import (
	"testing"
	"unsafe"

	"github.com/brickingsoft/brick/pkg/bytebuffer"
)

func TestAcquire(t *testing.T) {
	for i := 0; i < 10; i++ {
		buf := bytebuffer.Acquire()
		p0 := unsafe.SliceData(buf.Peek())
		buf.Write([]byte("hello world123"))
		p1 := unsafe.SliceData(buf.Peek())
		buf.Reset()
		buf.Write([]byte("hello world"))
		p2 := unsafe.SliceData(buf.Peek())
		p3 := unsafe.SliceData(buf.Peek()[0:])
		t.Log(p0, p1, p2, p3)
		bytebuffer.Release(buf)
	}
}
