package httpfs

import (
	"io/fs"

	"github.com/spf13/afero"
)

type aferoFS struct {
	fs afero.Fs
}

type aferoFile struct {
	f afero.File
}

var _ fs.FS = (*aferoFS)(nil)
var _ fs.File = (*aferoFile)(nil)

func NewAferoFS(a afero.Fs) fs.FS {
	return &aferoFS{fs: a}
}

func (a *aferoFS) Open(name string) (fs.File, error) {
	f, err := a.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return &aferoFile{f: f}, nil
}

func (a *aferoFile) Stat() (fs.FileInfo, error) {
	return a.f.Stat()
}

func (a *aferoFile) Read(p []byte) (int, error) {
	return a.f.Read(p)
}

func (a *aferoFile) Close() error {
	return a.f.Close()
}
