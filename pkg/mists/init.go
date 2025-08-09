package mists

import (
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
)

func init() {
	yaml.RegisterCustomMarshaler[Raw](func(raw Raw) ([]byte, error) {
		if len(raw) == 0 {
			return []byte{'{', '}'}, nil
		}
		if _, err := parser.ParseBytes(raw, 0); err != nil {
			return nil, err
		}
		return raw, nil
	})
	yaml.RegisterCustomUnmarshaler[Raw](func(r *Raw, b []byte) error {
		if len(b) == 0 {
			*r = []byte{'{', '}'}
			return nil
		}
		if _, err := parser.ParseBytes(b, 0); err != nil {
			return err
		}
		*r = b
		return nil
	})
}
