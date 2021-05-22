package httpfs

import (
	"io/fs"
	"net/http"

	"github.com/spf13/afero"
)

type aferoFS struct {
	fs afero.Fs
}

type aferoFile struct {
	f afero.File
}

var _ http.FileSystem = (*aferoFS)(nil)
var _ http.File = (*aferoFile)(nil)

func NewAferoFS(a afero.Fs) http.FileSystem {
	return &aferoFS{fs: a}
}

func (a *aferoFS) Open(name string) (http.File, error) {
	f, err := a.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return &aferoFile{f: f}, nil
}

func (a *aferoFile) Close() error {
	return a.f.Close()
}

func (a *aferoFile) Read(p []byte) (int, error) {
	return a.f.Read(p)
}

func (a *aferoFile) Seek(offset int64, whence int) (int64, error) {
	return a.f.Seek(offset, whence)
}

func (a *aferoFile) Readdir(count int) ([]fs.FileInfo, error) {
	return a.f.Readdir(count)
}

func (a *aferoFile) Stat() (fs.FileInfo, error) {
	return a.f.Stat()
}
