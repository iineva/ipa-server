package ipa

import (
	"path/filepath"
	"time"

	"github.com/iineva/ipa-server/pkg/plist"
)

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
	original   plist.Plist `json:"-"`
}

type AppList []*AppInfo

func (a AppList) Len() int           { return len(a) }
func (a AppList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AppList) Less(i, j int) bool { return a[i].Date.After(a[j].Date) }

func (a *AppInfo) IconStorageName() string {
	if a.NoneIcon {
		return ""
	}
	return filepath.Join(a.Identifier, a.ID, "icon.png")
}

func (a *AppInfo) IpaStorageName() string {
	return filepath.Join(a.Identifier, a.ID, "ipa.ipa")
}
