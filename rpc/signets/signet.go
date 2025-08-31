package signets

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"github.com/brickingsoft/brick/rpc/configs"
	"github.com/brickingsoft/brick/rpc/logs"
	"github.com/cespare/xxhash/v2"
)

type Signet interface {
	Print(b []byte) (signature []byte)
	Verify(b []byte, signature []byte) bool
}

type Options struct {
	Logger logs.Logger
	Hash   HashBuilder
	Config configs.Config
}

type Builder func(ctx context.Context, options Options) (Signet, error)

type HashBuilder func() hash.Hash

var XXHash = func() hash.Hash {
	return xxhash.New()
}

var SHA = func() hash.Hash { return sha1.New() }

var SHA256 = func() hash.Hash { return sha256.New() }

var SHA512 = func() hash.Hash { return sha512.New() }

var MD5 = func() hash.Hash { return md5.New() }
