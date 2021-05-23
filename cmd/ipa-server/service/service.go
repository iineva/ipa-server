package service

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"sort"
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
	MigrateOldData(list []*ipa.AppInfo) error
	Plist(id, publicURL string) ([]byte, error)
}

type service struct {
	list      ipa.AppList
	lock      sync.RWMutex
	store     storager.Storager
	publicURL string
}

func New(store storager.Storager, publicURL string) Service {
	return &service{
		store:     store,
		list:      ipa.AppList{},
		publicURL: publicURL, // use set public url
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
		item := s.itemInfo(row, publicURL)
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

	item := s.itemInfo(app, publicURL)
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
	var app *ipa.AppInfo
	for i, a := range s.list {
		if a.ID == id {
			app = a
			s.list = append(s.list[:i], s.list[i+1:]...)
			break
		}
	}
	s.lock.Unlock()

	if app == nil {
		return ErrIdNotFound
	}

	if err := s.store.Delete(app.IpaStorageName()); err != nil {
		return err
	}
	if err := s.store.Delete(app.IconStorageName()); err != nil {
		return err
	}
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

func (s *service) MigrateOldData(list []*ipa.AppInfo) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.list = append(list, s.list...)
	sort.Sort(s.list)
	return nil
}

func (s *service) Plist(id, publicURL string) ([]byte, error) {
	app, err := s.Find(id, publicURL)
	if err != nil {
		return nil, err
	}
	return NewInstallPlist(app)
}

func (s *service) find(id string) (*ipa.AppInfo, error) {
	for _, row := range s.list {
		if row.ID == id {
			return row, nil
		}
	}
	return nil, ErrIdNotFound
}

// get public url
func (s *service) storagerPublicURL(publicURL, name string) string {
	if s.publicURL != "" {
		publicURL = s.publicURL
	}
	u, err := s.store.PublicURL(publicURL, name)
	if err != nil {
		// TODO: handle err
		return ""
	}
	return u
}

func (s *service) servicePublicURL(publicURL, name string) string {
	if s.publicURL != "" {
		publicURL = s.publicURL
	}
	u, err := url.Parse(publicURL)
	if err != nil {
		// TODO: handle err
		return ""
	}
	u.Path = filepath.Join(u.Path, name)
	return u.String()
}

func (s *service) itemInfo(row *ipa.AppInfo, publicURL string) *Item {
	return &Item{
		AppInfo: *row,
		Ipa:     s.storagerPublicURL(publicURL, row.IpaStorageName()),
		Icon:    s.storagerPublicURL(publicURL, iconPath(row)),
		Plist:   s.servicePublicURL(publicURL, fmt.Sprintf("plist/%v.plist", row.ID)),
		WebIcon: s.storagerPublicURL(publicURL, iconPath(row)),
		Date:    time.Now(),
	}
}

func (s *service) history(row *ipa.AppInfo, publicURL string) []*Item {
	list := []*Item{}
	for _, i := range s.list {
		if i.Identifier == row.Identifier {
			item := s.itemInfo(i, publicURL)
			item.Current = i.ID == row.ID
			list = append(list, item)
		}
	}
	return list
}

func iconPath(app *ipa.AppInfo) string {
	name := app.IconStorageName()
	if name == "" {
		name = "img/default.png"
	}
	return name
}

type ServiceMiddleware func(Service) Service
