package ipa

import (
	"fmt"
	"time"
)

type IPAList []*IPA

// Item to use on web interface
type Item struct {
	IPA

	Ipa string `json:"ipa"`
	// Icon to display on iOS desktop
	Icon string `json:"icon"`
	// Plist to install ipa
	Plist string `json:"plist"`
	// WebIcon to display on web
	WebIcon string `json:"webIcon"`
	// Date
	Date time.Time `json:"date"`

	Current bool   `json:"current"`
	History []Item `json:"history"`
}

func (l *IPAList) Add() {

}

func iconPath(row *IPA) string {
	if row.NoneIcon {
		return "img/default.png"
	}
	return fmt.Sprintf("%s/%s/icon.png", row.Identifier, row.ID)
}

func (l *IPAList) List(publicURL string) []Item {
	list := []Item{}
	for _, row := range *l {
		has := false
		for _, i := range list {
			if i.Identifier == row.Identifier {
				has = true
				break
			}
		}
		if has {
			continue
		}
		item := l.itemInfo(row, publicURL)
		item.History = l.history(row, publicURL)
		list = append(list, item)
	}
	return list
}

func (l *IPAList) Find(id string, publicURL string) Item {
	var app *IPA
	for _, row := range *l {
		if row.ID == id {
			app = row
			break
		}
	}

	if app == nil {
		return Item{}
	}

	item := l.itemInfo(app, publicURL)
	item.History = l.history(app, publicURL)
	return item
}

func (l *IPAList) itemInfo(row *IPA, publicURL string) Item {
	return Item{
		IPA:     *row,
		Ipa:     fmt.Sprintf("%s/%s/%s/ipa.ipa", publicURL, row.Identifier, row.ID),
		Icon:    fmt.Sprintf("%s/%s", publicURL, iconPath(row)),
		Plist:   fmt.Sprintf("%s/plist/%v.plist", publicURL, row.ID),
		WebIcon: fmt.Sprintf("%s/%s", publicURL, iconPath(row)),
		Date:    time.Now(),
	}
}

func (l *IPAList) history(row *IPA, publicURL string) []Item {
	list := []Item{}
	for _, i := range *l {
		if i.Identifier == row.Identifier {
			item := l.itemInfo(i, publicURL)
			item.Current = i.ID == row.ID
			list = append(list, item)
		}
	}
	return list
}
