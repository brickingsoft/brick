package signets_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/brickingsoft/brick/pkg/signets"
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
