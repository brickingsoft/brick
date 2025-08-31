package bpack_test

import (
	"testing"

	"github.com/brickingsoft/brick/rpc/transports/bpack"
)

func TestDictionary(t *testing.T) {
	var hfs = []bpack.HeaderField{
		// f1
		{"f1", nil},
		{"f1", []byte("f1-a")},
		{"f1", []byte("f1-b")},
		{"f1", []byte("f1-c")},
		{"f1", []byte("f1-d")},
		{"f1", []byte("f1-b")},
		// f2
		{"f2", []byte("f2-a")},
		{"f2", []byte("f2-b")},
		{"f2", []byte("f2-c")},
		// f3
		{"f3", nil},
	}
	dict := &bpack.Dictionary{}
	dict.Load(hfs)
	hfs = append(hfs, bpack.HeaderField{Name: "f4"})

	for _, hf := range hfs {
		hv := hf.Value
		if len(hv) == 0 {
			hv = []byte("foo")
		}
		hi, vi := dict.Index([]byte(hf.Name), hv)
		if hi < 0 {
			t.Log("not found", hf.Name, string(hv))
			continue
		}
		if vi < 0 {
			name, value := dict.Get(hi)
			t.Log("found name", hf.Name, string(hv), hi, vi, string(name) == hf.Name, len(value) == 0)
			continue
		}
		name, value := dict.Get(vi)
		t.Log("found all", hf.Name, string(hv), hi, vi, string(name) == hf.Name, string(value) == string(hv))
	}
}
