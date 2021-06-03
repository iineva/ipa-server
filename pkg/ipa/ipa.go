package ipa

import (
	"archive/zip"
	"errors"
	"image"
	"image/png"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/iineva/CgbiPngFix/ipaPng"

	"github.com/iineva/bom/pkg/asset"
	"github.com/iineva/ipa-server/pkg/plist"
	"github.com/iineva/ipa-server/pkg/seekbuf"
)

var (
	ErrInfoPlistNotFound = errors.New("Info.plist not found")
)

const (
	// Payload/UnicornApp.app/AppIcon_TikTok76x76@2x~ipad.png
	// Payload/UnicornApp.app/AppIcon76x76.png
	newIconRegular   = `^Payload\/.*\.app\/AppIcon-?_?\w*(\d+(\.\d+)?)x(\d+(\.\d+)?)(@\dx)?(~ipad)?\.png$`
	oldIconRegular   = `^Payload\/.*\.app\/Icon-?_?\w*(\d+(\.\d+)?)?.png$`
	assetRegular     = `^Payload\/.*\.app/Assets.car$`
	infoPlistRegular = `^Payload\/.*\.app/Info.plist$`
)

// TODO: use InfoPlistIcon to parse icon files
type InfoPlistIcon struct {
	CFBundlePrimaryIcon struct {
		CFBundleIconFiles []string `json:"CFBundleIconFiles,omitempty"`
		CFBundleIconName  string   `json:"CFBundleIconName,omitempty"`
	} `json:"CFBundlePrimaryIcon,omitempty"`
}
type InfoPlist struct {
	CFBundleDisplayName        string        `json:"CFBundleDisplayName,omitempty"`
	CFBundleExecutable         string        `json:"CFBundleExecutable,omitempty"`
	CFBundleIconName           string        `json:"CFBundleIconName,omitempty"`
	CFBundleIcons              InfoPlistIcon `json:"CFBundleIcons,omitempty"`
	CFBundleIconsIpad          InfoPlistIcon `json:"CFBundleIcons~ipad,omitempty"`
	CFBundleIdentifier         string        `json:"CFBundleIdentifier,omitempty"`
	CFBundleName               string        `json:"CFBundleName,omitempty"`
	CFBundleShortVersionString string        `json:"CFBundleShortVersionString,omitempty"`
	CFBundleSupportedPlatforms []string      `json:"CFBundleSupportedPlatforms,omitempty"`
	CFBundleVersion            string        `json:"CFBundleVersion,omitempty"`
	// not standard
	Channel string `json:"channel"`
}

func Parse(readerAt io.ReaderAt, size int64) (*IPA, error) {

	r, err := zip.NewReader(readerAt, size)
	if err != nil {
		return nil, err
	}

	// match files
	var plistFile *zip.File
	var iconFiles []*zip.File
	var assetFile *zip.File
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
		if match, err = regexp.MatchString(oldIconRegular, f.Name); err != nil {
			return nil, err
		} else if match {
			iconFiles = append(iconFiles, f)
		}

		// parse new icons
		if match, err = regexp.MatchString(newIconRegular, f.Name); err != nil {
			return nil, err
		} else if match {
			iconFiles = append(iconFiles, f)
		}

		// parse Assets.car
		if match, err = regexp.MatchString(assetRegular, f.Name); err != nil {
			return nil, err
		} else if match {
			assetFile = f
		}

	}

	// parse Info.plist
	if plistFile == nil {
		return nil, ErrInfoPlistNotFound
	}
	var app *IPA
	{
		pf, err := plistFile.Open()
		defer pf.Close()
		if err != nil {
			return nil, err
		}
		info := &InfoPlist{}
		err = plist.Decode(pf, info)
		if err != nil {
			return nil, err
		}
		app = &IPA{
			info: info,
			size: size,
		}
	}

	// select bigest icon file
	var iconFile *zip.File
	var maxSize = -1
	for _, f := range iconFiles {
		size, err := iconSize(f.Name)
		if err != nil {
			return nil, err
		}
		if size > maxSize {
			maxSize = size
			iconFile = f
		}
	}
	// parse icon
	img, err := parseIconImage(iconFile)
	if err == nil {
		app.icon = img
	} else if assetFile != nil {
		// try get icon from Assets.car
		img, _ := parseIconAssets(assetFile)
		app.icon = img
	}

	return app, nil
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

func parseIconImage(iconFile *zip.File) (image.Image, error) {

	if iconFile == nil {
		return nil, errors.New("icon file is nil")
	}

	f, err := iconFile.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf, err := seekbuf.Open(f, seekbuf.MemoryMode)
	if err != nil {
		return nil, err
	}
	defer buf.Close()

	img, err := png.Decode(buf)
	if err != nil {
		// try fix to std png
		cgbi, err := ipaPng.Decode(buf)
		if err != nil {
			return nil, err
		}
		img = cgbi.Img
	}

	return img, nil
}

func parseIconAssets(assetFile *zip.File) (image.Image, error) {

	f, err := assetFile.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf, err := seekbuf.Open(f, seekbuf.MemoryMode)
	if err != nil {
		return nil, err
	}
	defer buf.Close()

	a, err := asset.NewWithReadSeeker(buf)
	if err != nil {
		return nil, err
	}
	return a.Image("AppIcon")
}
