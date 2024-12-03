package storager

import (
	"bufio"
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
	WRITER_BUFFER_SIZE   = 1024 * 1024 * 2 // 1M
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
	defer func() {
		_ = fi.Close()
	}()
	if err != nil {
		return err
	}

	// write with buffer
	w := bufio.NewWriterSize(fi, WRITER_BUFFER_SIZE)
	_, err = io.Copy(w, reader)
	if err != nil {
		return err
	}

	err = w.Flush()
	if err != nil {
		return err
	}

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
