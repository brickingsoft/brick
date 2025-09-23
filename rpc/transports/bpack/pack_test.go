package bpack_test

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/brickingsoft/brick/rpc/transports/bpack"
)

type Header struct {
	value http.Header
}

func (h *Header) Set(key []byte, value []byte) error {
	h.value.Set(string(key), string(value))
	return nil
}

func (h *Header) Iterator() bpack.HeaderIterator {
	return bpack.HeaderIterator(func(yield func(name []byte, value []byte) bool) {
		for ns, vss := range h.value {
			ns = strings.ToLower(ns)
			for _, vs := range vss {
				if !yield([]byte(ns), []byte(vs)) {
					return
				}
			}
		}
	})
}

func TestPack(t *testing.T) {
	pack, packErr := bpack.New(
		bpack.MaxHeaderSize(4096),
		bpack.Field("f1", "f1-a", "f1-b", "f1-c"),
		bpack.Field("f1", "f1-d", "f1-f"),
		bpack.Field("f2", "f2-a", "f2-b", "f2-c"),
		bpack.Field("f3"),
	)
	if packErr != nil {
		t.Fatal(packErr)
	}

	buf := bytes.NewBuffer(nil)

	h := &Header{
		value: make(http.Header),
	}

	h.Set([]byte("f1"), []byte("f1-a"))
	h.Set([]byte("f2"), []byte("f2-b"))
	h.Set([]byte("f3"), []byte("f3-x"))
	h.Set([]byte("fx"), []byte("fx-x"))

	wErr := pack.PackTo(buf, h.Iterator())
	if wErr != nil {
		t.Fatal(wErr)
	}

	t.Log("encode:", buf.Len()) // encode: 18

	h2 := &Header{
		value: make(http.Header),
	}

	rErr := pack.UnpackFrom(buf, h2)
	if rErr != nil {
		t.Fatal(rErr)
	}

	for name, value := range h2.Iterator() {
		t.Log("name:", string(name), "value:", string(value))
	}

}

func TestPack_DumpTo(t *testing.T) {
	src, packErr := bpack.New(
		bpack.MaxHeaderSize(4096),
		bpack.Field("f1", "f1-a", "f1-b", "f1-c"),
		bpack.Field("f1", "f1-d", "f1-f"),
		bpack.Field("f2", "f2-a", "f2-b", "f2-c"),
		bpack.Field("f3"),
	)
	if packErr != nil {
		t.Fatal(packErr)
	}

	buf := bytes.NewBuffer(nil)

	dErr := src.DumpTo(buf)
	if dErr != nil {
		t.Fatal(dErr)
	}

	pack := bpack.Acquire()
	defer bpack.Release(pack)

	loadErr := pack.LoadFrom(buf)
	if loadErr != nil {
		t.Fatal(loadErr)
	}

	buf.Reset()

	h := &Header{
		value: make(http.Header),
	}

	h.Set([]byte("f1"), []byte("f1-a"))
	h.Set([]byte("f2"), []byte("f2-b"))
	h.Set([]byte("f3"), []byte("f3-x"))
	h.Set([]byte("fx"), []byte("fx-x"))

	wErr := pack.PackTo(buf, h.Iterator())
	if wErr != nil {
		t.Fatal(wErr)
	}

	t.Log("encode:", buf.Len())

	h2 := &Header{
		value: make(http.Header),
	}

	rErr := pack.UnpackFrom(buf, h2)
	if rErr != nil {
		t.Fatal(rErr)
	}

	for name, value := range h2.Iterator() {
		t.Log("name:", string(name), "value:", string(value))
	}
}
