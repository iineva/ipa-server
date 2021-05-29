package apk

import (
	"io"

	"github.com/shogo82148/androidbinary"
	"github.com/shogo82148/androidbinary/apk"
)

func Parse(readerAt io.ReaderAt, size int64) (*APK, error) {
	pkg, err := apk.OpenZipReader(readerAt, size)
	if err != nil {
		return nil, err
	}
	defer pkg.Close()

	icon, err := pkg.Icon(&androidbinary.ResTableConfig{
		Density: 720,
	})
	if err != nil {
		// NOTE: ignore error
	}

	return &APK{
		icon:     icon,
		manifest: pkg.Manifest(),
		size:     size,
	}, nil
}
