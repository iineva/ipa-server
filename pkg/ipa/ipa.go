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

	"github.com/iineva/CgbiPngFix/ipaPng"

	"github.com/iineva/ipa-server/pkg/common"
	"github.com/iineva/ipa-server/pkg/plist"
	"github.com/iineva/ipa-server/pkg/seekbuf"
	"github.com/iineva/ipa-server/pkg/storager"
	"github.com/iineva/ipa-server/pkg/uuid"
)

type Reader interface {
	io.ReaderAt
	io.Reader
	io.Seeker
	Size() int64
}

var (
	ErrInfoPlistNotFound = errors.New("Info.plist not found")
)

const (
	tempDir = ".ipa_parser_temp"

	// Payload/UnicornApp.app/AppIcon_TikTok76x76@2x~ipad.png
	// Payload/UnicornApp.app/AppIcon76x76.png
	newIconRegular   = `^Payload\/.*\.app\/AppIcon-?_?\w*(\d+(\.\d+)?)x(\d+(\.\d+)?)(@\dx)?(~ipad)?\.png$`
	oldIconRegular   = `^Payload\/.*\.app\/Icon-?_?\w*(\d+(\.\d+)?)?.png$`
	infoPlistRegular = `^Payload\/.*\.app/Info.plist$`
)

type ParsedFiles struct {
	Icon  string
	Ipa   string
	Plist string
}

func ParseAndStorageIPA(readerAt Reader, store storager.Storager) (*AppInfo, *ParsedFiles, error) {

	// save ipa file
	parsedFiles := &ParsedFiles{}
	parsedFiles.Ipa = filepath.Join(tempDir, uuid.NewString())
	if err := store.Save(parsedFiles.Ipa, readerAt); err != nil {
		return nil, nil, err
	}

	readerAt.Seek(0, 0)
	r, err := zip.NewReader(readerAt, readerAt.Size())
	if err != nil {
		return nil, nil, err
	}

	// match files
	var plistFile *zip.File
	var iconFiles []*zip.File
	for _, f := range r.File {

		// parse Info.plist
		match, err := regexp.MatchString(infoPlistRegular, f.Name)
		{
			if err != nil {
				return nil, nil, err
			}
			if match {
				plistFile = f
			}
		}

		// parse old icons
		match, err = regexp.MatchString(oldIconRegular, f.Name)
		{
			if err != nil {
				return nil, nil, err
			}
			if match {
				iconFiles = append(iconFiles, f)
			}
		}

		// parse new icons
		match, _ = regexp.MatchString(newIconRegular, f.Name)
		{
			if err != nil {
				return nil, nil, err
			}
			if match {
				iconFiles = append(iconFiles, f)
			}
		}
	}

	if plistFile == nil {
		return nil, nil, ErrInfoPlistNotFound
	}

	// select bigest icon file
	var iconFile *zip.File
	var maxSize = -1
	for _, f := range iconFiles {
		size, err := iconSize(f.Name)
		if err != nil {
			return nil, nil, err
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
			return nil, nil, err
		}
		b, err := ioutil.ReadAll(pf)
		if err != nil {
			return nil, nil, err
		}

		info, err := plist.Parse(bytes.NewReader(b))
		if err != nil {
			return nil, nil, err
		}
		app = &AppInfo{
			ID:         uuid.NewString(),
			Name:       common.Def(info.GetString("CFBundleDisplayName"), info.GetString("CFBundleName"), info.GetString("CFBundleExecutable")),
			Version:    info.GetString("CFBundleShortVersionString"),
			Identifier: info.GetString("CFBundleIdentifier"),
			Build:      info.GetString("CFBundleVersion"),
			Channel:    info.GetString("channel"),
			Date:       time.Now(),
			Size:       readerAt.Size(),
			NoneIcon:   iconFile == nil,
			Metadata:   info,
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
			parsedFiles.Icon = filepath.Join(tempDir, uuid.NewString())
			if err := store.Save(parsedFiles.Icon, pngInput); err != nil {
				return nil, nil, err
			}
		}
	}

	return app, parsedFiles, nil
}

func iconSize(fileName string) (s int, err error) {
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
