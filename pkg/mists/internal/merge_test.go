package internal_test

import (
	"io"
	"testing"

	"github.com/brickingsoft/brick/pkg/mists/internal"
	"github.com/goccy/go-yaml/parser"
)

func TestMergeNode(t *testing.T) {
	var (
		dst = `a:
 s: a
 n: 1
 sub:
  s: sa
  n: 11
  ss:
   - s1
   - s2
b:
 s: b
 n: 2
c:
 - 1
 - 2
 - 3`
		src = `a:
 s: aa
 n: 11
 sub:
  b: true
  ss:
   - s3
   - s4
d:
 s: d
c:
 - 11
 - 22`
	)

	dstFile, dstErr := parser.ParseBytes([]byte(dst), 0)
	if dstErr != nil {
		t.Fatal(dstErr)
	}
	srcFile, srcErr := parser.ParseBytes([]byte(src), 0)
	if srcErr != nil {
		t.Fatal(srcErr)
	}
	err := internal.MergeNode(dstFile.Docs[0].Body, srcFile.Docs[0].Body)
	if err != nil {
		t.Fatal(err)
	}

	b, bErr := io.ReadAll(dstFile.Docs[0].Body)
	if bErr != nil {
		t.Fatal(bErr)
	}
	t.Log(string(b))

}
