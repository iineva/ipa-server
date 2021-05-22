package ipa

import (
	"archive/zip"
	"bytes"
	"errors"
	"image/png"
	"io"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lithammer/shortuuid"
	"github.com/poolqa/CgbiPngFix/ipaPng"

	"github.com/iineva/ipa-server/pkg/common"
	"github.com/iineva/ipa-server/pkg/plist"
	"github.com/iineva/ipa-server/pkg/seekbuf"
	"github.com/iineva/ipa-server/pkg/storager"
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

type IpaReader interface {
	io.ReaderAt
	io.Reader
	io.Seeker
}

var (
	ErrInfoPlistNotFound = errors.New("Info.plist not found")
)

const (
	// Payload/UnicornApp.app/AppIcon_TikTok76x76@2x~ipad.png
	// Payload/UnicornApp.app/AppIcon76x76.png
	newIconRegular   = `^Payload\/.*\.app\/AppIcon-?_?\w*(\d+(\.\d+)?)x(\d+(\.\d+)?)(@\dx)?(~ipad)?\.png$`
	oldIconRegular   = `^Payload\/.*\.app\/Icon-?_?\w*(\d+(\.\d+)?)?.png$`
	infoPlistRegular = `^Payload\/.*\.app/Info.plist$`
)

func ParseAndStorageIPA(readerAt IpaReader, size int64, store storager.Storager) (*AppInfo, error) {

	r, err := zip.NewReader(readerAt, size)
	if err != nil {
		return nil, err
	}

	// match files
	var plistFile *zip.File
	var iconFiles []*zip.File
	for _, f := range r.File {

		// parse Info.plist
		match, err := regexp.MatchString(infoPlistRegular, f.Name)
		{
			if err != nil {
				return nil, err
			}
			if match {
				plistFile = f
			}
		}

		// parse old icons
		match, err = regexp.MatchString(oldIconRegular, f.Name)
		{
			if err != nil {
				return nil, err
			}
			if match {
				iconFiles = append(iconFiles, f)
			}
		}

		// parse new icons
		match, _ = regexp.MatchString(newIconRegular, f.Name)
		{
			if err != nil {
				return nil, err
			}
			if match {
				iconFiles = append(iconFiles, f)
			}
		}
	}

	if plistFile == nil {
		return nil, ErrInfoPlistNotFound
	}

	// select bigest icon file
	var iconFile *zip.File
	var maxSize = -1
	for _, f := range iconFiles {
		size, err := IconSize(f.Name)
		if err != nil {
			return nil, err
		}
		if size > maxSize {
			maxSize = size
			iconFile = f
		}
	}

	// parse Info.plist
	var app *AppInfo
	{
		pf, err := plistFile.Open()
		defer pf.Close()
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(pf)
		if err != nil {
			return nil, err
		}

		info, err := plist.Parse(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		app = &AppInfo{
			ID:         NewAppID(),
			Name:       common.Def(info.GetString("CFBundleDisplayName"), info.GetString("CFBundleName"), info.GetString("CFBundleExecutable")),
			Version:    info.GetString("CFBundleShortVersionString"),
			Identifier: info.GetString("CFBundleIdentifier"),
			Build:      info.GetString("CFBundleVersion"),
			Channel:    info.GetString("channel"),
			Date:       time.Now(),
			Size:       size,
			NoneIcon:   iconFile == nil,
			original:   info,
		}
	}

	if iconFile != nil {
		// try fix png for browser
		f, err := iconFile.Open()
		defer f.Close()
		buf, _ := seekbuf.Open(f, seekbuf.MemoryMode)
		defer buf.Close()
		var pngInput io.Reader = buf
		if err == nil {
			if err == nil {
				cgbi, err := ipaPng.Decode(buf)
				if err == nil {
					b := bytes.NewBuffer(make([]byte, 0))
					err = png.Encode(b, cgbi.Img)
					if err == nil {
						// if png fix done, reset pngInput
						pngInput = b
					}
				}
			}

			// save icon file
			buf.Seek(0, 0)
			if err := store.Save(pngInput, filepath.Join(app.Identifier, app.ID, "icon.png")); err != nil {
				return nil, err
			}
		}
	}

	// save ipa file
	readerAt.Seek(0, 0)
	if err := store.Save(readerAt, filepath.Join(app.Identifier, app.ID, "ipa.ipa")); err != nil {
		return nil, err
	}

	return app, nil
}

func IconSize(fileName string) (s int, err error) {
	size := float64(0)
	match, _ := regexp.MatchString(oldIconRegular, fileName)
	name := strings.TrimSuffix(filepath.Base(fileName), ".png")
	if match {
		arr := strings.Split(name, "-")
		if len(arr) == 2 {
			size, err = strconv.ParseFloat(arr[1], 32)
		} else {
			size = 160
		}
	}
	match, _ = regexp.MatchString(newIconRegular, fileName)
	if match {
		s := strings.Split(name, "@")[0]
		s = strings.Split(s, "x")[1]
		s = strings.Split(s, "~")[0]
		size, err = strconv.ParseFloat(s, 32)
		if strings.Index(name, "@2x") != -1 {
			size *= 2
		} else if strings.Index(name, "@3x") != -1 {
			size *= 3
		}
	}
	return int(size), err
}

func NewAppID() string {
	return shortuuid.New()
}
