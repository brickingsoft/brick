package signets

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"hash"
	"sync"
)

func ECDSA(keyPEM []byte, builder HashBuilder) (Signet, error) {
	if len(keyPEM) == 0 {
		return nil, errors.Join(errors.New("failed to create ecdsa signet"), errors.New("key is missing"))
	}
	if builder == nil {
		return nil, errors.Join(errors.New("failed to create ecdsa signet"), errors.New("hash builder is missing"))
	}
	block, _ := pem.Decode(keyPEM)
	privateKey, priErr := x509.ParsePKCS8PrivateKey(block.Bytes)
	if priErr != nil {
		return nil, errors.Join(errors.New("failed to create ecdsa signet"), priErr)
	}
	key, ok := privateKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.Join(errors.New("failed to create ecdsa signet"), errors.New("invalid private key type"))
	}
	return &ECDSASignet{
		pub:     &key.PublicKey,
		pri:     key,
		builder: builder,
		pool:    sync.Pool{},
	}, nil
}

type ECDSASignet struct {
	pub     *ecdsa.PublicKey
	pri     *ecdsa.PrivateKey
	builder HashBuilder
	pool    sync.Pool
}

func (s *ECDSASignet) acquireHash() hash.Hash {
	v := s.pool.Get()
	if v == nil {
		return s.builder()
	}
	return v.(hash.Hash)
}

func (s *ECDSASignet) releaseHash(hash hash.Hash) {
	hash.Reset()
	s.pool.Put(hash)
}

func (s *ECDSASignet) Print(b []byte) (signature []byte) {
	h := s.acquireHash()
	h.Write(b)
	p := h.Sum(nil)
	v, err := ecdsa.SignASN1(rand.Reader, s.pri, p)
	if err != nil {
		panic(err)
	}
	signature = make([]byte, base64.URLEncoding.EncodedLen(len(v)))
	base64.URLEncoding.Encode(signature, v)
	s.releaseHash(h)
	return
}

func (s *ECDSASignet) Verify(b []byte, signature []byte) bool {
	dst := make([]byte, base64.URLEncoding.DecodedLen(len(signature)))
	n, err := base64.URLEncoding.Decode(dst, signature)
	if err != nil {
		return false
	}
	dst = dst[:n]
	h := s.acquireHash()
	h.Write(b)
	p := h.Sum(nil)
	ok := ecdsa.VerifyASN1(s.pub, p, dst)
	s.releaseHash(h)
	return ok
}
