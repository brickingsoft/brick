package mem

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

func NewDir(name string) (*Dir, error) {
	if name == "" {
		return nil, os.ErrInvalid
	}
	return &Dir{
		name:    name,
		modTime: time.Now(),
		entries: nil,
	}, nil
}

type Entry interface {
	Name() string
	IsDir() bool
	ModTime() time.Time
	Info() fs.FileInfo
}

type DirEntry struct {
	value Entry
}

func (entry *DirEntry) Name() string {
	return entry.value.Name()
}

func (entry *DirEntry) IsDir() bool {
	return entry.value.IsDir()
}

func (entry *DirEntry) Type() fs.FileMode {
	return fs.ModeDir
}

func (entry *DirEntry) Info() (fs.FileInfo, error) {
	return entry.value.Info(), nil
}

type Dir struct {
	name    string
	modTime time.Time
	entries []fs.DirEntry
}

func (d *Dir) Sub(dir string) (fs.FS, error) {
	if dir == "" {
		return nil, os.ErrInvalid
	}
	for _, entry := range d.entries {
		if entry.Name() == dir {
			v := entry.(*DirEntry)
			if entry.IsDir() {
				return v.value.(*Dir), nil
			} else {
				return nil, os.ErrInvalid
			}
		}
	}
	return nil, os.ErrNotExist
}

func (d *Dir) Stat() (fs.FileInfo, error) {
	info := d.Info()
	return info, nil
}

func (d *Dir) Read(_ []byte) (int, error) {
	return 0, errors.New("can not read a directory")
}

func (d *Dir) Close() error {
	return nil
}

func (d *Dir) Name() string {
	return d.name
}

func (d *Dir) IsDir() bool {
	return true
}

func (d *Dir) ModTime() time.Time {
	return d.modTime
}

func (d *Dir) Info() fs.FileInfo {
	return &FileInfo{
		name:    d.name,
		size:    int64(len(d.entries)),
		modTime: d.modTime,
		isDir:   true,
	}
}

func (d *Dir) AddFile(name string, b []byte) error {
	file, fileErr := NewFile(name, b)
	if fileErr != nil {
		return fileErr
	}
	return d.Add(file)
}

func (d *Dir) Add(file *File) error {
	if file == nil {
		return os.ErrInvalid
	}
	d.entries = append(d.entries, &DirEntry{
		value: file,
	})
	return nil
}

func (d *Dir) AddDir(child *Dir) error {
	if child == nil {
		return fs.ErrInvalid
	}
	d.entries = append(d.entries, &DirEntry{
		value: child,
	})
	return nil
}

func (d *Dir) Open(name string) (fs.File, error) {
	if len(d.entries) == 0 {
		return nil, fs.ErrNotExist
	}
	name = filepath.Base(name)
	if name == "." {
		return d, nil
	}
	for _, entry := range d.entries {
		if entry.Name() == name {
			v := entry.(*DirEntry)
			if entry.IsDir() {
				dir := v.value.(*Dir)
				return dir, nil
			}
			file := v.value.(*File)
			return file, nil
		}
	}
	return nil, fs.ErrNotExist
}

func (d *Dir) ReadDir(name string) ([]fs.DirEntry, error) {
	if len(d.entries) == 0 {
		return nil, fs.ErrNotExist
	}
	name = filepath.Base(name)
	if name == "." {
		return d.entries, nil
	}
	for _, entry := range d.entries {
		if entry.Name() == name {
			if entry.IsDir() {
				v := entry.(*DirEntry)
				dir := v.value.(*Dir)
				return dir.entries, nil
			}
			return nil, fs.ErrInvalid
		}
	}
	return nil, fs.ErrNotExist
}
