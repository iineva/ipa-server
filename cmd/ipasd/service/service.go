package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/iineva/ipa-server/pkg/apk"
	"github.com/iineva/ipa-server/pkg/ipa"
	"github.com/iineva/ipa-server/pkg/storager"
	"github.com/iineva/ipa-server/pkg/uuid"
)

var (
	ErrIdNotFound = errors.New("id not found")
)

const (
	tempDir = ".ipa_parser_temp"
)

// Item to use on web interface
type Item struct {
	// from AppInfo
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Date       time.Time `json:"date"`
	Size       int64     `json:"size"`
	Channel    string    `json:"channel"`
	Build      string    `json:"build"`
	Version    string    `json:"version"`
	Identifier string    `json:"identifier"`

	// package download link
	Pkg string `json:"pkg"`
	// Icon to display on iOS desktop
	Icon string `json:"icon"`
	// Plist to install ipa
	Plist string `json:"plist,omitempty"`
	// WebIcon to display on web
	WebIcon string `json:"webIcon"`
	// Type 0:ios 1:android
	Type AppInfoType `json:"type"`

	Current bool    `json:"current"`
	History []*Item `json:"history,omitempty"`
}

func (i *Item) String() string {
	return fmt.Sprintf("%+v", *i)
}

type Service interface {
	List(publicURL string) ([]*Item, error)
	Find(id string, publicURL string) (*Item, error)
	History(id string, publicURL string) ([]*Item, error)
	Delete(id string) error
	Add(r Reader, t AppInfoType) error
	Plist(id, publicURL string) ([]byte, error)
}

type Reader interface {
	io.Reader
	io.ReaderAt
	Size() int64
}

type service struct {
	list         AppList
	lock         sync.RWMutex
	store        storager.Storager
	publicURL    string
	metadataName string
}

func New(store storager.Storager, publicURL, metadataName string) Service {
	s := &service{
		store:        store,
		list:         AppList{},
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
	var app *AppInfo
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

	if err := s.store.Delete(app.PackageStorageName()); err != nil {
		return err
	}
	if !app.NoneIcon {
		if err := s.store.Delete(app.IconStorageName()); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) Add(r Reader, t AppInfoType) error {

	app, err := s.addPackage(r, t)
	if err != nil {
		return err
	}

	// update list
	s.lock.Lock()
	s.list = append([]*AppInfo{app}, s.list...)
	s.lock.Unlock()

	return s.saveMetadata()
}

func (s *service) addPackage(r Reader, t AppInfoType) (*AppInfo, error) {
	// save ipa file to temp
	pkgTempFileName := filepath.Join(tempDir, uuid.NewString())
	if err := s.store.Save(pkgTempFileName, r); err != nil {
		return nil, err
	}

	// parse package
	var pkg Package
	var err error
	switch t {
	case AppInfoTypeIpa:
		pkg, err = ipa.Parse(r, r.Size())
	case AppInfoTypeApk:
		pkg, err = apk.Parse(r, r.Size())
	}
	if err != nil {
		return nil, err
	}

	// new AppInfo
	app := NewAppInfo(pkg, t)
	if err != nil {
		return nil, err
	}
	// move temp package file to target location
	err = s.store.Move(pkgTempFileName, app.PackageStorageName())
	if err != nil {
		return nil, err
	}

	// try save icon file
	if pkg.Icon() != nil {
		buf := &bytes.Buffer{}
		err = png.Encode(buf, pkg.Icon())
		if err == nil {
			if err := s.store.Save(app.IconStorageName(), buf); err != nil {
				// NOTE: ignore error
			}
		}
	}

	return app, nil
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

	list := AppList{}
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

func (s *service) find(id string) (*AppInfo, error) {
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

func (s *service) itemInfo(row *AppInfo, publicURL string) *Item {

	plist := ""
	switch row.Type {
	case AppInfoTypeIpa:
		plist = s.servicePublicURL(publicURL, fmt.Sprintf("plist/%v.plist", row.ID))
	}

	return &Item{
		// from AppInfo
		ID:         row.ID,
		Name:       row.Name,
		Date:       row.Date,
		Size:       row.Size,
		Build:      row.Build,
		Identifier: row.Identifier,
		Version:    row.Version,
		Channel:    row.Channel,
		Type:       row.Type,

		Pkg:     s.storagerPublicURL(publicURL, row.PackageStorageName()),
		Plist:   plist,
		Icon:    s.iconPublicURL(publicURL, row),
		WebIcon: s.iconPublicURL(publicURL, row),
	}
}

func (s *service) history(row *AppInfo, publicURL string) []*Item {
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

func (s *service) iconPublicURL(publicURL string, app *AppInfo) string {
	name := app.IconStorageName()
	if name == "" {
		name = "img/default.png"
		return s.servicePublicURL(publicURL, name)
	}
	return s.storagerPublicURL(publicURL, name)
}

type ServiceMiddleware func(Service) Service
