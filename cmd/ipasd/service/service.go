package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/iineva/ipa-server/pkg/ipa"
	"github.com/iineva/ipa-server/pkg/storager"
)

// Item to use on web interface
type Item struct {
	// from ipa.AppInfo
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Date       time.Time `json:"date"`
	Size       int64     `json:"size"`
	Channel    string    `json:"channel"`
	Build      string    `json:"build"`
	Version    string    `json:"version"`
	Identifier string    `json:"identifier"`

	Ipa string `json:"ipa"`
	// Icon to display on iOS desktop
	Icon string `json:"icon"`
	// Plist to install ipa
	Plist string `json:"plist"`
	// WebIcon to display on web
	WebIcon string `json:"webIcon"`

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
	Add(r ipa.Reader, size int64) error
	Plist(id, publicURL string) ([]byte, error)
}

type service struct {
	list         ipa.AppList
	lock         sync.RWMutex
	store        storager.Storager
	publicURL    string
	metadataName string
}

func New(store storager.Storager, publicURL, metadataName string) Service {
	s := &service{
		store:        store,
		list:         ipa.AppList{},
		publicURL:    publicURL, // use set public url
		metadataName: metadataName,
	}
	if err := s.tryMigrateOldData(); err != nil {
		// NOTE: ignore error
	}
	return s
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

	if err := s.saveMetadata(); err != nil {
		return err
	}

	if err := s.store.Delete(app.IpaStorageName()); err != nil {
		return err
	}
	if !app.NoneIcon {
		if err := s.store.Delete(app.IconStorageName()); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) Add(r ipa.Reader, size int64) error {

	// parse and save ipa
	app, err := ipa.ParseAndStorageIPA(r, size, s.store)
	if err != nil {
		return err
	}

	// update list
	s.lock.Lock()
	s.list = append([]*ipa.AppInfo{app}, s.list...)
	s.lock.Unlock()

	return s.saveMetadata()
}

// save metadata
func (s *service) saveMetadata() error {
	s.lock.Lock()
	d, err := json.Marshal(s.list)
	s.lock.Unlock()

	if err != nil {
		return err
	}

	b := bytes.NewBuffer(d)
	return s.store.Save(s.metadataName, b)
}

func (s *service) tryMigrateOldData() error {
	f, err := s.store.OpenMetadata(s.metadataName)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	list := ipa.AppList{}
	if err := json.Unmarshal(b, &list); err != nil {
		return err
	}

	s.lock.Lock()
	s.list = append(list, s.list...)
	sort.Sort(s.list)
	s.lock.Unlock()

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
		// from ipa.AppInfo
		ID:         row.ID,
		Name:       row.Name,
		Date:       row.Date,
		Size:       row.Size,
		Build:      row.Build,
		Identifier: row.Identifier,
		Version:    row.Version,
		Channel:    row.Channel,

		Ipa:     s.storagerPublicURL(publicURL, row.IpaStorageName()),
		Icon:    s.iconPublicURL(publicURL, row),
		Plist:   s.servicePublicURL(publicURL, fmt.Sprintf("plist/%v.plist", row.ID)),
		WebIcon: s.iconPublicURL(publicURL, row),
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

func (s *service) iconPublicURL(publicURL string, app *ipa.AppInfo) string {
	name := app.IconStorageName()
	if name == "" {
		name = "img/default.png"
		return s.servicePublicURL(publicURL, name)
	}
	return s.storagerPublicURL(publicURL, name)
}

type ServiceMiddleware func(Service) Service
