package signets

import (
	"crypto/hmac"
	"encoding/base64"
	"errors"
	"hash"
	"sync"
)

func HMAC(key []byte, builder HashBuilder) (Signet, error) {
	if len(key) == 0 {
		return nil, errors.Join(errors.New("failed to create hmac signet"), errors.New("key is missing"))
	}
	if builder == nil {
		return nil, errors.Join(errors.New("failed to create hmac signet"), errors.New("hash builder is missing"))
	}
	return &HMACSignet{
		key:     key,
		builder: builder,
		pool:    sync.Pool{},
	}, nil
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
	signature = make([]byte, base64.URLEncoding.EncodedLen(len(v)))
	base64.URLEncoding.Encode(signature, v)
	s.releaseHash(h)
	return
}

func (s *HMACSignet) Verify(b []byte, signature []byte) bool {
	dst := make([]byte, base64.URLEncoding.DecodedLen(len(signature)))
	n, err := base64.URLEncoding.Decode(dst, signature)
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
