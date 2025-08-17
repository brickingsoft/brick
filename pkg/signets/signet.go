package signets

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"github.com/cespare/xxhash/v2"
)

type Signet interface {
	Print(b []byte) (signature []byte)
	Verify(b []byte, signature []byte) bool
}

type HashBuilder func() hash.Hash

var XXHash = func() hash.Hash {
	return xxhash.New()
}

var SHA = func() hash.Hash { return sha1.New() }

var SHA256 = func() hash.Hash { return sha256.New() }

var SHA512 = func() hash.Hash { return sha512.New() }

var MD5 = func() hash.Hash { return md5.New() }
