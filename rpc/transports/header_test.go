package transports_test

import (
	"testing"

	"github.com/brickingsoft/brick/rpc/transports"
	"github.com/quic-go/quic-go/quicvarint"
)

func TestHeader(t *testing.T) {
	h := transports.AcquireHeader()
	defer transports.ReleaseHeader(h)

	h.SetAgent([]byte("agent"), []byte("device"))
	h.AddForwarded([]byte("name"), []byte("host"), []byte("proto"))
	h.SetAuthorization([]byte("authorization"))
	h.SetContentLength(10)
	h.SetContentType([]byte("Content-Type"))
	h.SetContentEncoding([]byte("Content-Encoding"))

	setErr := h.Set([]byte("for"), []byte("bar"))
	if setErr != nil {
		t.Fatal(setErr)
	}

	for name, value := range h.Iterator() {
		if string(name) == transports.ContentLengthHeaderStringKey {
			n, _, parseErr := quicvarint.Parse(value)
			t.Log(string(name), n, parseErr)
			continue
		}
		t.Log(string(name), string(value))
	}
}
