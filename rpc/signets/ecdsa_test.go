package signets_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/brickingsoft/brick/rpc/configs"
	"github.com/brickingsoft/brick/rpc/signets"
)

func createECDSASignets() (v signets.Signet, err error) {
	key, keyErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if keyErr != nil {
		return v, keyErr
	}
	pkcs8, pkcs8Err := x509.MarshalPKCS8PrivateKey(key)
	if pkcs8Err != nil {
		return v, pkcs8Err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   pkcs8,
	})
	tempFile, tempFileErr := os.CreateTemp("", "brick-signet-ecdsa-*.pem")
	if tempFileErr != nil {
		return v, tempFileErr
	}

	_, wErr := tempFile.Write(keyPEM)
	if wErr != nil {
		return v, wErr
	}
	keyPath := tempFile.Name()
	_ = tempFile.Close()

	defer os.Remove(keyPath)
	config, configErr := configs.NewConfig([]byte(fmt.Sprintf("key_path: \"%s\"", filepath.ToSlash(keyPath))))
	if configErr != nil {
		return v, configErr
	}
	options := signets.Options{
		Logger: nil,
		Config: config,
	}
	_, builder := signets.ECDSA()
	v, err = builder(options)
	return
}

func TestECDSA(t *testing.T) {
	signet, err := createECDSASignets()
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
	signet, err := createECDSASignets()
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
