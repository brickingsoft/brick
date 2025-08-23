package mem_test

import (
	"fmt"
	"io/fs"
	"testing"

	"github.com/brickingsoft/brick/pkg/fs/mem"
)

func TestNewDir(t *testing.T) {
	dir, dirErr := mem.NewDir("config.d")
	if dirErr != nil {
		t.Fatal(dirErr)
	}
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("%d.txt", i)
		content := []byte(name)
		fileErr := dir.AddFile(name, content)
		if fileErr != nil {
			t.Fatal(fileErr)
		}
	}
	child, childErr := mem.NewDir("child.d")
	if childErr != nil {
		t.Fatal(childErr)
	}
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("child-%d.txt", i)
		content := []byte(name)
		fileErr := child.AddFile(name, content)
		if fileErr != nil {
			t.Fatal(fileErr)
		}
	}
	_ = dir.AddDir(child)

	readFull(t, dir)

}

func readFull(t *testing.T, dir fs.FS) {
	entries, readErr := fs.ReadDir(dir, ".")
	if readErr != nil {
		t.Fatal(readErr)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			sub, subErr := fs.Sub(dir, entry.Name())
			if subErr != nil {
				t.Fatal(subErr)
			}
			readFull(t, sub)
			continue
		}
		b, bErr := fs.ReadFile(dir, entry.Name())
		if bErr != nil {
			t.Fatal(entry.Name(), bErr)
		}
		t.Log(entry.Name(), string(b), entry.Name() == string(b))
	}
}
