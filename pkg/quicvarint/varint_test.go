package quicvarint_test

import (
	"bytes"
	"testing"

	"github.com/brickingsoft/brick/pkg/quicvarint"
)

func TestWrite(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	wn, wErr := quicvarint.Write(buf, 10)
	if wErr != nil {
		t.Fatal(wErr)
	}
	t.Log("write >", wn)

	n, rErr := quicvarint.Read(buf)
	if rErr != nil {
		t.Fatal(rErr)
	}
	t.Log("read >", n)
}
