package service

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/iineva/ipa-server/pkg/ipa"
)

// Item to use on web interface
type Item struct {
	ipa.IPA

	Ipa string `json:"ipa"`
	// Icon to display on iOS desktop
	Icon string `json:"icon"`
	// Plist to install ipa
	Plist string `json:"plist"`
	// WebIcon to display on web
	WebIcon string `json:"webIcon"`
	// Date
	Date time.Time `json:"date"`

	Current bool    `json:"current"`
	History []*Item `json:"history,omitempty"`
}

func (i *Item) String() string {
	return fmt.Sprintf("%+v", *i)
}

var (
	ErrIdNotFound = errors.New("id not found")
)

type Service interface {
	List(publicURL string) ([]*Item, error)
	Find(id string, publicURL string) (*Item, error)
	History(id string, publicURL string) ([]*Item, error)
	Delete(id string) error
	Add(io.Reader) error
}

type service struct {
	list []*ipa.IPA
}

func New() Service {
	return &service{
		list: []*ipa.IPA{{
			ID:         "364f38fb19c247eaa7f3fb1e8d5fc79e",
			Name:       "Test",
			Version:    "1.0",
			Identifier: "com.ineva.test",
			Build:      "129",
			Channel:    "",
			Date:       time.Now(),
			Size:       1204 * 1024 * 10,
			NoneIcon:   true,
		}, {
			ID:         "215b468df5a04d3a8fb22d4478c88f1d",
			Name:       "Test",
			Version:    "1.0",
			Identifier: "com.ineva.test",
			Build:      "128",
			Channel:    "",
			Date:       time.Now(),
			Size:       1204 * 1024,
			NoneIcon:   true,
		}},
	}
}

func (s *service) List(publicURL string) ([]*Item, error) {
	list := []*Item{}
	for _, row := range s.list {
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
		item := itemInfo(row, publicURL)
		item.History = s.history(row, publicURL)
		list = append(list, item)
	}
	return list, nil
}

func (s *service) Find(id string, publicURL string) (*Item, error) {
	app, err := s.find(id)
	if err != nil {
		return nil, err
	}

	item := itemInfo(app, publicURL)
	item.History = s.history(app, publicURL)
	return item, nil
}

func (s *service) History(id string, publicURL string) ([]*Item, error) {
	app, err := s.find(id)
	if err != nil {
		return nil, err
	}
	return s.history(app, publicURL), nil
}

func (s *service) Delete(id string) error {
	// TODO:
	return nil
}
func (s *service) Add(io.Reader) error {
	// TODO:
	return nil
}

func (s *service) find(id string) (*ipa.IPA, error) {
	for _, row := range s.list {
		if row.ID == id {
			return row, nil
		}
	}
	return nil, ErrIdNotFound
}

func itemInfo(row *ipa.IPA, publicURL string) *Item {
	return &Item{
		IPA:     *row,
		Ipa:     fmt.Sprintf("%s/%s/%s/ipa.ipa", publicURL, row.Identifier, row.ID),
		Icon:    fmt.Sprintf("%s/%s", publicURL, iconPath(row)),
		Plist:   fmt.Sprintf("%s/plist/%v.plist", publicURL, row.ID),
		WebIcon: fmt.Sprintf("%s/%s", publicURL, iconPath(row)),
		Date:    time.Now(),
	}
}

func (s *service) history(row *ipa.IPA, publicURL string) []*Item {
	list := []*Item{}
	for _, i := range s.list {
		if i.Identifier == row.Identifier {
			item := itemInfo(i, publicURL)
			item.Current = i.ID == row.ID
			list = append(list, item)
		}
	}
	return list
}

func iconPath(row *ipa.IPA) string {
	if row.NoneIcon {
		return "img/default.png"
	}
	return fmt.Sprintf("%s/%s/icon.png", row.Identifier, row.ID)
}

type ServiceMiddleware func(Service) Service
