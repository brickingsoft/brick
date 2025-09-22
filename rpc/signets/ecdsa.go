package signets

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"hash"
	"os"
	"strings"
	"sync"
)

type ECDSASignetConfig struct {
	KeyPath string `yaml:"key_path"`
	Hash    string `yaml:"hash"`
}

func ECDSA() (string, Builder) {
	return "ecdsa", func(options Options) (v Signet, err error) {
		config := ECDSASignetConfig{}
		if err = options.Config.As(&config); err != nil {
			err = errors.Join(errors.New("failed to create ecdsa signet"), err)
			return
		}
		keyPath := strings.TrimSpace(config.KeyPath)
		if keyPath == "" {
			err = errors.Join(errors.New("failed to create ecdsa signet"), errors.New("key is missing"))
			return
		}
		keyPEM, keyErr := os.ReadFile(keyPath)
		if keyErr != nil {
			err = errors.Join(errors.New("failed to create ecdsa signet"), keyErr)
			return
		}
		block, _ := pem.Decode(keyPEM)
		privateKey, priErr := x509.ParsePKCS8PrivateKey(block.Bytes)
		if priErr != nil {
			err = errors.Join(errors.New("failed to create ecdsa signet"), priErr)
			return
		}
		key, ok := privateKey.(*ecdsa.PrivateKey)
		if !ok {
			err = errors.Join(errors.New("failed to create ecdsa signet"), errors.New("invalid private key type"))
			return
		}
		hash0 := strings.TrimSpace(config.Hash)
		if hash0 == "" {
			hash0 = "xxhash"
		}
		hb := getHashBuilder(hash0)
		if hb == nil {
			err = errors.Join(errors.New("failed to create ecdsa signet"), errors.New("hash builder not found"))
			return
		}

		v = &ECDSASignet{
			pub:     &key.PublicKey,
			pri:     key,
			builder: hb,
			pool:    sync.Pool{},
		}
		return
	}
}

func ECDSA1(keyPEM []byte, builder HashBuilder) (Signet, error) {
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
	signature = make([]byte, hex.EncodedLen(len(v)))
	hex.Encode(signature, v)
	s.releaseHash(h)
	return
}

func (s *ECDSASignet) Verify(b []byte, signature []byte) bool {
	dst := make([]byte, hex.DecodedLen(len(signature)))
	n, err := hex.Decode(dst, signature)
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
