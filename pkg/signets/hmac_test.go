package signets_test

import (
	"testing"

	"github.com/brickingsoft/brick/pkg/signets"
)

func TestHMAC(t *testing.T) {
	key := []byte("secret_key")
	signet, err := signets.HMAC(key, signets.XXHash)
	if err != nil {
		t.Fatal(err)
	}
	b := []byte("hello world")
	p := signet.Print(b)
	t.Log(string(p))
	t.Log(signet.Verify(b, p))
}
