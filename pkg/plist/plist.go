package plist

import (
	"io"

	"howett.net/plist"

	"github.com/iineva/ipa-server/pkg/seekbuf"
)

func Decode(r io.Reader, d interface{}) error {
	buf, err := seekbuf.Open(r, seekbuf.MemoryMode)
	if err != nil {
		return err
	}
	if err := plist.NewDecoder(buf).Decode(d); err != nil {
		return err
	}
	return nil
}
