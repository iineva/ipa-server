package apk

import (
	"fmt"
	"image"

	"github.com/shogo82148/androidbinary/apk"
)

type APK struct {
	manifest apk.Manifest
	icon     image.Image
	size     int64
}

func (a *APK) Name() string {
	return a.manifest.App.Label.MustString()
}

func (a *APK) Version() string {
	return a.manifest.VersionName.MustString()
}

func (a *APK) Identifier() string {
	return a.manifest.Package.MustString()
}

func (a *APK) Build() string {
	return fmt.Sprintf("%v", a.manifest.VersionCode.MustInt32())
}

func (a *APK) Channel() string {
	for _, r := range a.manifest.App.MetaData {
		n := r.Name.MustString()
		if n == "channel" {
			return r.Value.MustString()
		}
	}
	return ""
}

func (a *APK) Icon() image.Image {
	return a.icon
}

func (a *APK) Size() int64 {
	return a.size
}
