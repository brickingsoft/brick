package mists

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/brickingsoft/brick/pkg/mists/internal"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

func RegisterCustomMarshaler[T any](marshaler func(T) ([]byte, error)) {
	yaml.RegisterCustomMarshaler[T](marshaler)
}

func RegisterCustomUnmarshaler[T any](unmarshaler func(r *T, b []byte) error) {
	yaml.RegisterCustomUnmarshaler[T](unmarshaler)
}

func Valid(b []byte) bool {
	if _, err := parser.ParseBytes(b, 0); err != nil {
		return false
	}
	return true
}

func Empty() (mist *Mist) {
	mist = &Mist{
		raw: []byte{'{', '}'},
	}
	return
}

// New
//
// create Mist from yaml
func New(b []byte) (mist *Mist, err error) {
	if len(b) == 0 {
		mist = Empty()
		return
	}
	file, parseErr := parser.ParseBytes(b, 0)
	if parseErr != nil {
		err = parseErr
		return
	}
	if len(file.Docs) == 0 {
		mist = Empty()
		return
	}
	body := file.Docs[0].Body
	if body == nil {
		mist = Empty()
		return
	}
	anchors, anchorsErr := internal.Anchors(body)
	if anchorsErr != nil {
		err = anchorsErr
		return
	}

	if len(anchors) > 0 {
		ctx := context.TODO()
		replaceErr := internal.AliasReplaceAnchor(ctx, body, anchors)
		if replaceErr != nil {
			err = replaceErr
			return
		}
	}
	bodyBytes, readErr := io.ReadAll(body)
	if readErr != nil {
		if errors.Is(readErr, io.EOF) {
			mist = Empty()
			return
		}
		err = readErr
		return
	}
	mist = &Mist{
		raw: bodyBytes,
	}
	return
}

type Mist struct {
	raw []byte
}

func (cfg *Mist) Unmarshal(v any) (err error) {
	err = yaml.Unmarshal(cfg.raw, v)
	return
}

func (cfg *Mist) Bytes() (b []byte) {
	b = make([]byte, len(cfg.raw))
	copy(b, cfg.raw)
	return
}

func (cfg *Mist) String() string {
	return string(cfg.raw)
}

func (cfg *Mist) Node(name string) (target *Mist, has bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	file, parseErr := parser.ParseBytes(cfg.raw, 0)
	if parseErr != nil {
		return
	}
	for _, doc := range file.Docs {
		if doc.Type() != ast.MappingType {
			continue
		}
		node, ok := doc.Body.(*ast.MappingNode)
		if !ok {
			continue
		}
		iter := node.MapRange()
		for iter.Next() {
			key := iter.Key()
			keyNode, isString := key.(*ast.StringNode)
			if !isString {
				continue
			}
			if keyNode.Value != name {
				continue
			}
			value := iter.Value()
			b, bErr := io.ReadAll(value)
			if bErr != nil {
				return
			}
			target = &Mist{
				raw: b,
			}
			has = true
			return
		}
	}

	return
}

func (cfg *Mist) Range(iter func(item *Mist) (stop bool)) (err error) {
	file, parseErr := parser.ParseBytes(cfg.raw, 0)
	if parseErr != nil {
		err = parseErr
		return
	}
	for _, doc := range file.Docs {
		switch doc.Type() {
		case ast.SequenceType:
			node, ok := doc.Body.(*ast.SequenceNode)
			if !ok {
				continue
			}
			for _, value := range node.Values {
				b, bErr := io.ReadAll(value)
				if bErr != nil {
					err = bErr
					return
				}
				if iter(&Mist{raw: b}) {
					return
				}
			}
			break
		default:
			continue
		}
	}
	return
}

func (cfg *Mist) Path(expr string) (target *Mist, err error) {
	path, pathErr := yaml.PathString(expr)
	if pathErr != nil {
		err = pathErr
		return
	}
	node, nodeErr := path.ReadNode(bytes.NewBuffer(cfg.raw))
	if nodeErr != nil {
		err = nodeErr
		return
	}
	b, bErr := io.ReadAll(node)
	if bErr != nil {
		err = bErr
		return
	}
	target = &Mist{
		raw: b,
	}
	return
}

func (cfg *Mist) Merge(target *Mist) (err error) {
	if target == nil {
		return
	}
	src, srcErr := parser.ParseBytes(target.Bytes(), 0)
	if srcErr != nil {
		err = srcErr
		return
	}
	if len(src.Docs) == 0 || src.Docs[0] == nil {
		return
	}
	dst, dstErr := parser.ParseBytes(cfg.Bytes(), 0)
	if dstErr != nil {
		err = dstErr
		return
	}
	if len(dst.Docs) == 0 || dst.Docs[0] == nil {
		cfg.raw = target.Bytes()
		return
	}
	dstBody := dst.Docs[0].Body
	srcBody := src.Docs[0].Body
	if err = internal.MergeNode(dstBody, srcBody); err != nil {
		return
	}
	b, bErr := io.ReadAll(dstBody)
	if bErr != nil {
		err = bErr
		return
	}
	cfg.raw = b
	return
}

func (cfg *Mist) Empty() bool {
	return len(cfg.raw) == 0 || bytes.Equal(cfg.raw, []byte{'{', '}'})
}
