package service

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/iineva/ipa-server/pkg/ipa"
	"github.com/iineva/ipa-server/pkg/seekbuf"
	"github.com/iineva/ipa-server/pkg/storager"
)

// Item to use on web interface
type Item struct {
	ipa.AppInfo

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
	Add(r io.Reader, size int64) error
}

type service struct {
	list  []*ipa.AppInfo
	lock  sync.RWMutex
	store storager.Storager
}

func New(store storager.Storager) Service {
	return &service{
		store: store,
		list:  []*ipa.AppInfo{},
	}
}

func (s *service) List(publicURL string) ([]*Item, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
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
	s.lock.RLock()
	defer s.lock.RUnlock()
	app, err := s.find(id)
	if err != nil {
		return nil, err
	}

	item := itemInfo(app, publicURL)
	item.History = s.history(app, publicURL)
	return item, nil
}

func (s *service) History(id string, publicURL string) ([]*Item, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	app, err := s.find(id)
	if err != nil {
		return nil, err
	}
	return s.history(app, publicURL), nil
}

func (s *service) Delete(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	// TODO:
	return nil
}

func (s *service) Add(r io.Reader, size int64) error {
	buf, err := seekbuf.Open(r, seekbuf.FileMode)
	defer buf.Close()
	if err != nil {
		return err
	}
	app, err := ipa.ParseAndStorageIPA(buf, size, s.store)
	if err != nil {
		return err
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.list = append([]*ipa.AppInfo{app}, s.list...)
	return nil
}

func (s *service) find(id string) (*ipa.AppInfo, error) {
	for _, row := range s.list {
		if row.ID == id {
			return row, nil
		}
	}
	return nil, ErrIdNotFound
}

func itemInfo(row *ipa.AppInfo, publicURL string) *Item {
	return &Item{
		AppInfo: *row,
		Ipa:     fmt.Sprintf("%s/%s/%s/ipa.ipa", publicURL, row.Identifier, row.ID),
		Icon:    fmt.Sprintf("%s/%s", publicURL, iconPath(row)),
		Plist:   fmt.Sprintf("%s/plist/%v.plist", publicURL, row.ID),
		WebIcon: fmt.Sprintf("%s/%s", publicURL, iconPath(row)),
		Date:    time.Now(),
	}
}

func (s *service) history(row *ipa.AppInfo, publicURL string) []*Item {
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

func iconPath(row *ipa.AppInfo) string {
	if row.NoneIcon {
		return "img/default.png"
	}
	return fmt.Sprintf("%s/%s/icon.png", row.Identifier, row.ID)
}

type ServiceMiddleware func(Service) Service
