package storager

import (
	"io"
)

type Storager interface {
	Save(name string, reader io.Reader) error
	Delete(name string) error
	PublicURL(publicURL, name string) (string, error)
}
