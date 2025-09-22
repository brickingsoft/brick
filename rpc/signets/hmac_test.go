package signets_test

import (
	"testing"

	"github.com/brickingsoft/brick/rpc/configs"
	"github.com/brickingsoft/brick/rpc/signets"
)

func TestHMAC(t *testing.T) {
	config, _ := configs.NewConfig([]byte(`key: "secret_key"`))
	options := signets.Options{
		Logger: nil,
		Config: config,
	}
	_, builder := signets.HMAC()
	signet, err := builder(options)
	if err != nil {
		t.Fatal(err)
	}
	b := []byte("hello world")
	p := signet.Print(b)
	t.Log(string(p))
	t.Log(signet.Verify(b, p))
}

// BenchmarkHMAC-20    	 8784438	       135.0 ns/op	      40 b/op	       4 allocs/op
func BenchmarkHMAC(b *testing.B) {
	config, _ := configs.NewConfig([]byte(`key: "secret_key"`))
	options := signets.Options{
		Logger: nil,
		Config: config,
	}
	_, builder := signets.HMAC()
	signet, err := builder(options)
	if err != nil {
		b.Fatal(err)
	}
	p := []byte("hello world")
	b.ReportAllocs()
	b.ResetTimer()
	n := 0.0
	for i := 0; i < b.N; i++ {
		s := signet.Print(p)
		if !signet.Verify(p, s) {
			n++
			b.ReportMetric(n, "failed")
		}
	}
}
