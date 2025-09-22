package signets

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"strings"

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
	Config *configs.Config
}

type Builder func(options Options) (Signet, error)

type NamedBuilder func() (name string, builder Builder)

type HashBuilder func() hash.Hash

var (
	hashBuilderMap = map[string]HashBuilder{
		"xxhash": func() hash.Hash { return xxhash.New() },
		"sha1":   func() hash.Hash { return sha1.New() },
		"sha256": func() hash.Hash { return sha256.New() },
		"sha512": func() hash.Hash { return sha512.New() },
		"md5":    func() hash.Hash { return md5.New() },
	}
)

func RegisterHashBuilder(name string, builder HashBuilder) {
	name = strings.TrimSpace(name)
	if name == "" || builder == nil {
		return
	}
	name = strings.ToLower(name)
	hashBuilderMap[name] = builder
}

func getHashBuilder(name string) (builder HashBuilder) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	name = strings.ToLower(name)
	builder = hashBuilderMap[name]
	return
}
