// Package multipart to handle MultipartForm
package multipart

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
)

type MultipartForm struct {
	r *http.Request
}

type FormFile struct {
	part     *multipart.Part
	name     string // form name
	filename string // file name
	// size     int64  // readed size
}

var _ io.Reader = (*FormFile)(nil)

func New(r *http.Request) *MultipartForm {
	return &MultipartForm{r: r}
}

func (m *MultipartForm) GetFormFile(targetName string) (*FormFile, error) {
	mr, err := m.multipartReader(false)
	if err != nil {
		return nil, err
	}

	p, err := mr.NextPart()
	if err != nil {
		return nil, err
	}

	name := p.FormName()
	if name != targetName {
		return nil, fmt.Errorf("want %s got %s", targetName, name)
	}
	filename := p.FileName()

	return &FormFile{
		part:     p,
		name:     name,
		filename: filename,
	}, nil
}

// code copy from http/request.go:447
func (m *MultipartForm) multipartReader(allowMixed bool) (*multipart.Reader, error) {
	r := m.r
	v := r.Header.Get("Content-Type")
	if v == "" {
		return nil, http.ErrNotMultipart
	}
	d, params, err := mime.ParseMediaType(v)
	if err != nil || !(d == "multipart/form-data" || allowMixed && d == "multipart/mixed") {
		return nil, http.ErrNotMultipart
	}
	boundary, ok := params["boundary"]
	if !ok {
		return nil, http.ErrMissingBoundary
	}
	return multipart.NewReader(r.Body, boundary), nil
}

func (f *FormFile) Read(p []byte) (n int, err error) {
	return f.part.Read(p)
}

func (f *FormFile) FileName() string {
	return f.filename
}

func (f *FormFile) Name() string {
	return f.name
}
