package internal_test

import (
	"testing"

	"github.com/brickingsoft/brick/pkg/configs/internal"
	"github.com/goccy/go-yaml/parser"
)

func TestAnchors(t *testing.T) {
	var (
		b = []byte(`
a: &a1
  s: "a1"
  n: 11
  t1: "t1"
b: &a2
  s: "a2"
  n: 22
  t2: "t2"
  <<: *a1
sub:
  x: "xxx"
  c: &a3
    s: "a3"
    n: 33
    t3: "t3"
    <<: *a2
`)
	)
	file, fileErr := parser.ParseBytes(b, 0)
	if fileErr != nil {
		t.Fatal(fileErr)
	}
	nodes, rErr := internal.Anchors(file.Docs[0].Body)
	if rErr != nil {
		t.Fatal(rErr)
	}
	for key, value := range nodes {
		t.Log(key)
		t.Log(value)
	}
}
