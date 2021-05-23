package storager

import (
	"io"
	"log"
	"net/url"
	"path/filepath"

	"github.com/spf13/afero"
)

type oferoStorager struct {
	fs afero.Fs
}

const (
	oferoStoragerDirPerm = 0755
)

var _ Storager = (*oferoStorager)(nil)

func NewAferoStorager(fs afero.Fs) Storager {
	return &oferoStorager{fs: fs}
}

func NewOsFileStorager(basepath string) Storager {
	return NewAferoStorager(afero.NewBasePathFs(afero.NewOsFs(), basepath))
}

func NewMemStorager() Storager {
	return NewAferoStorager(afero.NewMemMapFs())
}

func (f *oferoStorager) Save(name string, reader io.Reader) error {
	dir := filepath.Dir(name)
	if err := f.fs.MkdirAll(dir, oferoStoragerDirPerm); err != nil {
		return err
	}
	fi, err := f.fs.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(fi, reader)
	return err
}

func (f *oferoStorager) Delete(name string) error {
	return f.fs.Remove(name)
}

func (f *oferoStorager) PublicURL(publicURL, name string) (string, error) {
	log.Print(publicURL, " ", name)
	u, err := url.Parse(publicURL)
	if err != nil {
		return "", err
	}
	u.Path = filepath.Join(u.Path, name)
	return u.String(), nil
}
