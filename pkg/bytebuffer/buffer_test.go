package bytebuffer_test

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/brickingsoft/brick/pkg/bytebuffer"
)

func TestBuffer_Read(t *testing.T) {
	buf := bytebuffer.Acquire()
	defer bytebuffer.Release(buf)
	rb := make([]byte, 1024)
	rbn, _ := rand.Read(rb)
	_, _ = buf.Write(rb[:rbn])
	b, rErr := io.ReadAll(buf)
	_, _ = buf.Seek(0, io.SeekStart)
	t.Log(bytes.Equal(b, buf.CloneBytes()), rErr)
}
