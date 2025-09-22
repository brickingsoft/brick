package signets

import (
	"crypto/hmac"
	"encoding/hex"
	"errors"
	"hash"
	"strings"
	"sync"
)

type HMACSignetConfig struct {
	Key  string `yaml:"key"`
	Hash string `yaml:"hash"`
}

func HMAC() (string, Builder) {
	return "hmac", func(options Options) (v Signet, err error) {
		config := HMACSignetConfig{}
		if err = options.Config.As(&config); err != nil {
			err = errors.Join(errors.New("failed to create hmac signet"), err)
			return
		}
		key := strings.TrimSpace(config.Key)
		if key == "" {
			err = errors.Join(errors.New("failed to create hmac signet"), errors.New("key is missing"))
			return
		}
		hash0 := strings.TrimSpace(config.Hash)
		if hash0 == "" {
			hash0 = "xxhash"
		}
		hb := getHashBuilder(hash0)
		if hb == nil {
			err = errors.Join(errors.New("failed to create hmac signet"), errors.New("hash builder not found"))
			return
		}

		v = &HMACSignet{
			key:     []byte(key),
			builder: hb,
			pool:    sync.Pool{},
		}

		return
	}
}

type HMACSignet struct {
	key     []byte
	builder HashBuilder
	pool    sync.Pool
}

func (s *HMACSignet) acquireHash() hash.Hash {
	v := s.pool.Get()
	if v == nil {
		return hmac.New(s.builder, s.key)
	}
	return v.(hash.Hash)
}

func (s *HMACSignet) releaseHash(hash hash.Hash) {
	hash.Reset()
	s.pool.Put(hash)
}

func (s *HMACSignet) Print(b []byte) (signature []byte) {
	h := s.acquireHash()
	h.Write(b)
	v := h.Sum(nil)
	signature = make([]byte, hex.EncodedLen(len(v)))
	hex.Encode(signature, v)
	s.releaseHash(h)
	return
}

func (s *HMACSignet) Verify(b []byte, signature []byte) bool {
	dst := make([]byte, hex.DecodedLen(len(signature)))
	n, err := hex.Decode(dst, signature)
	if err != nil {
		return false
	}
	dst = dst[:n]
	h := s.acquireHash()
	h.Write(b)
	p := h.Sum(nil)
	ok := hmac.Equal(p, dst)
	s.releaseHash(h)
	return ok
}
