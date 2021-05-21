package ipa

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iineva/ipa-server/pkg/common"
	"github.com/iineva/ipa-server/pkg/plist"
)

type IPA struct {
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

func ReadPlist(readerAt io.ReaderAt, size int64) (*IPA, error) {

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
	// TODO: save icon file
	log.Printf("TODO: save icon file: %s", iconFile.Name)

	// parse Info.plist
	var ipa *IPA
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
		ipa = &IPA{
			ID:         strings.Replace(uuid.NewString(), "-", "", -1),
			Name:       common.Def(info.GetString("CFBundleDisplayName"), info.GetString("CFBundleName"), info.GetString("CFBundleExecutable")),
			Version:    info.GetString("CFBundleShortVersionString"),
			Identifier: info.GetString("CFBundleIdentifier"),
			Build:      info.GetString("CFBundleVersion"),
			Channel:    info.GetString("channel"),
			Date:       time.Now(),
			Size:       size,
			NoneIcon:   iconFile != nil,
			original:   info,
		}
	}

	return ipa, nil
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
