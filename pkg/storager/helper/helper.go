package helper

import (
	"io"
	"net/url"
	"path/filepath"
)

type CallbackAfterReaderClose struct {
	cb     func() error
	reader io.ReadCloser
}

func NewCallbackAfterReaderClose(reader io.ReadCloser, cb func() error) io.ReadCloser {
	return &CallbackAfterReaderClose{reader: reader, cb: cb}
}

func (d *CallbackAfterReaderClose) Close() error {
	if err := d.reader.Close(); err != nil {
		return err
	}
	return d.cb()
}

func (d *CallbackAfterReaderClose) Read(p []byte) (int, error) {
	return d.reader.Read(p)
}

func UrlJoin(u string, p string) (string, error) {
	d, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	d.Path = filepath.Join(d.Path, p)
	return d.String(), nil
}
