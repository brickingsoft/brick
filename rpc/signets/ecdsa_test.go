package signets_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/brickingsoft/brick/rpc/signets"
)

func TestECDSA(t *testing.T) {
	key, keyErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if keyErr != nil {
		t.Fatal(keyErr)
	}
	pkcs8, pkcs8Err := x509.MarshalPKCS8PrivateKey(key)
	if pkcs8Err != nil {
		t.Fatal(pkcs8Err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   pkcs8,
	})
	signet, err := signets.ECDSA(keyPEM, signets.XXHash)
	if err != nil {
		t.Fatal(err)
	}
	b := []byte("hello world")
	p := signet.Print(b)
	t.Log(string(p))
	t.Log(signet.Verify(b, p))
}

// BenchmarkECDSA-20    	   10000	    106743 ns/op	    7166 b/op	      81 allocs/op
func BenchmarkECDSA(b *testing.B) {
	key, keyErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if keyErr != nil {
		b.Fatal(keyErr)
	}
	pkcs8, pkcs8Err := x509.MarshalPKCS8PrivateKey(key)
	if pkcs8Err != nil {
		b.Fatal(pkcs8Err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   pkcs8,
	})
	signet, err := signets.ECDSA(keyPEM, signets.XXHash)
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
