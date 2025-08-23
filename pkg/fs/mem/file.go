package mem

import (
	"io"
	"io/fs"
	"os"
	"time"
)

func NewFile(name string, b []byte) (*File, error) {
	if name == "" {
		return nil, os.ErrInvalid
	}
	if len(b) == 0 {
		return nil, fs.ErrInvalid
	}
	return &File{
		name:    name,
		modTime: time.Now(),
		b:       b,
	}, nil
}

type File struct {
	name    string
	modTime time.Time
	b       []byte
}

func (file *File) Name() string {
	return file.name
}

func (file *File) IsDir() bool {
	return false
}

func (file *File) ModTime() time.Time {
	return file.modTime
}

func (file *File) Info() fs.FileInfo {
	return &FileInfo{
		name:    file.name,
		size:    int64(len(file.b)),
		modTime: file.modTime,
		isDir:   false,
	}
}

func (file *File) Stat() (fs.FileInfo, error) {
	info := file.Info()
	return info, nil
}

func (file *File) Size() int64 {
	return int64(len(file.b))
}

func (file *File) Read(b []byte) (int, error) {
	if len(file.b) == 0 {
		return 0, io.EOF
	}
	if len(b) == 0 {
		return 0, fs.ErrInvalid
	}
	n := copy(b, file.b)
	file.b = file.b[n:]
	return n, nil
}

func (file *File) Close() error {
	return nil
}
