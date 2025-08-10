package mists_test

import (
	"testing"

	"github.com/brickingsoft/brick/pkg/mists"
)

const (
	root = `default: &default
 s: b
 n: 2
a:
 s: a
 n: 1
 sub:
  s: sa
  n: 11
  ss:
   - s1
   - s2
b:
 n: 1
 <<: *default
c:
 - 1
 - 2
 - 3`
	target = `default: &default
 s: d
a:
 s: aa
 n: 11
 sub:
  b: true
  ss:
   - s3
   - s4
d:
 s: dd
 <<: *default
c:
 - 11
 - 22`
)

func TestNew(t *testing.T) {
	rootConfig, rootErr := mists.New([]byte(root))
	if rootErr != nil {
		t.Fatal(rootErr)
	}
	t.Log(rootConfig)
}

func TestMist_Merge(t *testing.T) {
	rootConfig, rootErr := mists.New([]byte(root))
	if rootErr != nil {
		t.Fatal(rootErr)
	}
	targetConfig, targetErr := mists.New([]byte(target))
	if targetErr != nil {
		t.Fatal(targetErr)
	}
	err := rootConfig.Merge(targetConfig)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(rootConfig.Bytes()))
}
