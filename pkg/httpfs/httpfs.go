// Package httpfs Merge multiple http.FileSystem as one
package httpfs

import (
	"net/http"
)

type httpFS struct {
	fss []http.FileSystem
}

var _ http.FileSystem = (*httpFS)(nil)

func New(fss ...http.FileSystem) http.FileSystem {
	return &httpFS{fss: fss}
}

func (h *httpFS) Open(name string) (f http.File, err error) {
	for _, cfs := range h.fss {
		f, err = cfs.Open(name)
		if err != nil {
			continue
		}
		return f, err
	}
	return nil, err
}
