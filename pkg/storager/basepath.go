package storager

import (
	"io"
	"path/filepath"
)

type basepathStorager struct {
	base string
	s    Storager
}

var _ Storager = (*basepathStorager)(nil)

func NewBasePathStorager(basepath string, store Storager) Storager {
	return &basepathStorager{base: basepath, s: store}
}

func (b *basepathStorager) Save(name string, reader io.Reader) error {
	return b.s.Save(filepath.Join(b.base, name), reader)
}

func (b *basepathStorager) OpenMetadata(name string) (io.ReadCloser, error) {
	return b.s.OpenMetadata(filepath.Join(b.base, name))
}

func (b *basepathStorager) Delete(name string) error {
	return b.s.Delete(filepath.Join(b.base, name))
}

func (b *basepathStorager) Move(src, dest string) error {
	return b.s.Move(filepath.Join(b.base, src), filepath.Join(b.base, dest))
}

func (b *basepathStorager) PublicURL(publicURL, name string) (string, error) {
	return b.s.PublicURL(publicURL, filepath.Join(b.base, name))
}
