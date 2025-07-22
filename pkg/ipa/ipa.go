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

	"github.com/iineva/bom/pkg/asset"
	"github.com/iineva/ipa-server/pkg/plist"
	"github.com/iineva/ipa-server/pkg/seekbuf"
	"github.com/poolqa/CgbiPngFix/ipaPng"
)

var (
	ErrInfoPlistNotFound = errors.New("Info.plist not found")
)

var (
	// Payload/UnicornApp.app/AppIcon_TikTok76x76@2x~ipad.png
	// Payload/UnicornApp.app/AppIcon76x76.png
	newIconRegular   = regexp.MustCompile(`^Payload\/.*\.app\/AppIcon-?_?\w*(\d+(\.\d+)?)x(\d+(\.\d+)?)(@\dx)?(~ipad)?\.png$`)
	oldIconRegular   = regexp.MustCompile(`^Payload\/.*\.app\/Icon-?_?\w*(\d+(\.\d+)?)?.png$`)
	assetRegular     = regexp.MustCompile(`^Payload\/.*\.app/Assets.car$`)
	infoPlistRegular = regexp.MustCompile(`^Payload\/.*\.app/Info.plist$`)
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
	// not standard
	ISMetaData map[string]interface{} `json:"ISMetaData,omitempty"`
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
		if infoPlistRegular.MatchString(f.Name) {
			plistFile = f
		}

		// parse old icons
		if oldIconRegular.MatchString(f.Name) {
			iconFiles = append(iconFiles, f)
		}

		// parse new icons
		if newIconRegular.MatchString(f.Name) {
			iconFiles = append(iconFiles, f)
		}

		// parse Assets.car
		if assetRegular.MatchString(f.Name) {
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
		if err != nil {
			return nil, err
		}
		defer pf.Close()
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
			continue
		}
		if size > maxSize {
			maxSize = size
			iconFile = f
		}
	}
	// if can't find bigest one, just first one.
	if iconFile == nil && len(iconFiles) > 0 {
		iconFile = iconFiles[0]
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
	name := strings.TrimSuffix(filepath.Base(fileName), ".png")
	if oldIconRegular.MatchString(fileName) {
		arr := strings.Split(name, "-")
		if len(arr) == 2 {
			size, err = strconv.ParseFloat(arr[1], 32)
		} else {
			size = 160
		}
	}
	if newIconRegular.MatchString(fileName) {
		s := strings.Split(name, "@")[0]
		s = strings.Split(s, "x")[1]
		s = strings.Split(s, "~")[0]
		size, err = strconv.ParseFloat(s, 32)
		if strings.Contains(name, "@2x") {
			size *= 2
		} else if strings.Contains(name, "@3x") {
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
		if _, err := buf.Seek(0, 0); err != nil {
			return nil, err
		}
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
