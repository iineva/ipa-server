package service

import (
	"image"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/iineva/ipa-server/pkg/uuid"
)

type AppInfoType int
type AppInfo struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Version    string      `json:"version"`
	Identifier string      `json:"identifier"`
	Build      string      `json:"build"`
	Channel    string      `json:"channel"`
	Date       time.Time   `json:"date"`
	Size       int64       `json:"size"`
	NoneIcon   bool        `json:"noneIcon"`
	Type       AppInfoType `json:"type"`
	// Metadata   plist.Plist `json:"metadata"` // metadata from Info.plist
}

const (
	AppInfoTypeIpa     = AppInfoType(0)
	AppInfoTypeApk     = AppInfoType(1)
	AppInfoTypeUnknown = AppInfoType(-1)
)

func (t AppInfoType) StorageName() string {
	switch t {
	case AppInfoTypeIpa:
		return "ipa.ipa"
	case AppInfoTypeApk:
		return "apk.apk"
	default:
		return "unknown"
	}
}

func FileType(n string) AppInfoType {
	ext := strings.ToLower(path.Ext(n))
	switch ext {
	case ".ipa":
		return AppInfoTypeIpa
	case ".apk":
		return AppInfoTypeApk
	default:
		return AppInfoTypeUnknown
	}
}

type AppList []*AppInfo

func (a AppList) Len() int           { return len(a) }
func (a AppList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AppList) Less(i, j int) bool { return a[i].Date.After(a[j].Date) }

type Package interface {
	Name() string
	Version() string
	Identifier() string
	Build() string
	Channel() string
	Icon() image.Image
	Size() int64
}

func NewAppInfo(i Package, t AppInfoType) *AppInfo {
	return &AppInfo{
		ID:         uuid.NewString(),
		Name:       i.Name(),
		Version:    i.Version(),
		Identifier: i.Identifier(),
		Build:      i.Build(),
		Channel:    i.Channel(),
		Date:       time.Now(),
		Size:       i.Size(),
		Type:       t,
		NoneIcon:   i.Icon() == nil,
	}
}

func (a *AppInfo) IconStorageName() string {
	if a.NoneIcon {
		return ""
	}
	return filepath.Join(a.Identifier, a.ID, "icon.png")
}

func (a *AppInfo) PackageStorageName() string {
	return filepath.Join(a.Identifier, a.ID, a.Type.StorageName())
}
