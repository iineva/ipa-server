package storager

import (
	"io"
)

type Storager interface {
	Save(name string, reader io.Reader) error
	OpenMetadata(name string) (io.ReadCloser, error)
	Delete(name string) error
	Move(src, dest string) error
	PublicURL(publicURL, name string) (string, error)
}
