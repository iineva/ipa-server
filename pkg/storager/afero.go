package storager

import (
	"io"
	"path/filepath"

	"github.com/iineva/ipa-server/pkg/storager/helper"
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

func (f *oferoStorager) OpenMetadata(name string) (io.ReadCloser, error) {
	return f.fs.Open(name)
}

func (f *oferoStorager) Delete(name string) error {
	err := f.fs.Remove(name)
	if err != nil {
		return err
	}
	// auto delete empty dir
	err = f.deleteEmptyDir(filepath.Dir(name))
	if err != nil {
		// NOTE: ignore error
	}
	return nil
}

func (f *oferoStorager) deleteEmptyDir(name string) error {
	name = filepath.Clean(name)
	if name == "." {
		return nil
	}

	err := f.fs.Remove(name)
	if err != nil {
		return err
	}

	return f.deleteEmptyDir(filepath.Dir(name))
}

func (f *oferoStorager) Move(src, dest string) error {
	err := f.fs.MkdirAll(filepath.Dir(dest), oferoStoragerDirPerm)
	if err != nil {
		return err
	}
	return f.fs.Rename(src, dest)
}

func (f *oferoStorager) PublicURL(publicURL, name string) (string, error) {
	return helper.UrlJoin(publicURL, name)
}
