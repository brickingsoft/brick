package mists

import (
	"context"
	"errors"
	"io"

	"github.com/brickingsoft/brick/pkg/mists/internal"
	"github.com/goccy/go-yaml/parser"
)

type Mist interface {
	Decode(v any) (err error)
	Raw() (b Raw)
	Node(name string) (target Mist, has bool)
	Range(iter func(item Mist) (stop bool)) (err error)
	Path(expr string) (target Mist, err error)
	Merge(target Mist) (err error)
}

func New(raw []byte) (mist Mist, err error) {
	if len(raw) == 0 {
		mist = &Config{
			raw: []byte{'{', '}'},
		}
		return
	}
	file, parseErr := parser.ParseBytes(raw, 0)
	if parseErr != nil {
		err = parseErr
		return
	}
	if len(file.Docs) == 0 {
		mist = &Config{
			raw: []byte{'{', '}'},
		}
		return
	}
	body := file.Docs[0].Body
	if body == nil {
		mist = &Config{
			raw: []byte{'{', '}'},
		}
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
			mist = &Config{
				raw: []byte{'{', '}'},
			}
			return
		}
		err = readErr
		return
	}
	mist = &Config{
		raw: bodyBytes,
	}
	return
}
