package ipa

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
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
	newIconRegular   = `^Payload\/.*\.app\/AppIcon-?(\d+(\.\d+)?)x(\d+(\.\d+)?)(@\dx)?.*\.png$`
	oldIconRegular   = `^Payload\/.*\.app\/Icon-?(\d+(\.\d+)?)?.png$`
	infoPlistRegular = `^Payload\/.*\.app/Info.plist$`
)

func ReadPlist(readerAt io.ReaderAt, size int64) (*IPA, error) {

	r, err := zip.NewReader(readerAt, size)
	if err != nil {
		return nil, err
	}

	// match files
	var plistFile, oldIconFile, newIconFile *zip.File
	for _, f := range r.File {
		if plistFile != nil && (oldIconFile != nil || newIconFile != nil) {
			break
		}

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
				oldIconFile = f
			}
		}

		// parse new icons
		match, _ = regexp.MatchString(newIconRegular, f.Name)
		{
			if err != nil {
				return nil, err
			}
			if match {
				newIconFile = f
			}
		}
	}

	if plistFile == nil {
		return nil, ErrInfoPlistNotFound
	}

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
			Name:       def(info.GetString("CFBundleDisplayName"), info.GetString("CFBundleName"), info.GetString("CFBundleExecutable")),
			Version:    info.GetString("CFBundleShortVersionString"),
			Identifier: info.GetString("CFBundleIdentifier"),
			Build:      info.GetString("CFBundleVersion"),
			Channel:    info.GetString("channel"),
			Date:       time.Now(),
			Size:       size,
			NoneIcon:   newIconFile == nil && oldIconFile == nil,
			original:   info,
		}
	}

	return ipa, nil
}

// get args until arg is not empty
func def(args ...string) string {
	for _, v := range args {
		if v != "" {
			return v
		}
	}
	return ""
}
