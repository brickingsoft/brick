package mem

import (
	"io/fs"
	"time"
)

type FileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (info *FileInfo) Name() string {
	return info.name
}

func (info *FileInfo) Size() int64 {
	return info.size
}

func (info *FileInfo) Mode() fs.FileMode {
	if info.isDir {
		return fs.ModeDir
	}
	return fs.ModeIrregular
}

func (info *FileInfo) ModTime() time.Time {
	return info.modTime
}

func (info *FileInfo) IsDir() bool {
	return info.isDir
}

func (info *FileInfo) Sys() any {
	return nil
}
