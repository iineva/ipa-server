package storager

import "io"

type Storager interface {
	Save(reader io.Reader, name string) error
	Delete(name string) error
	PublicURL(name string) (string, error)
}
