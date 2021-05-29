package service

import (
	"bytes"
	"io/ioutil"
	"text/template"
)

const plistTpl = `<?xml version="1.0" encoding="utf-8"?>
<plist version="1.0">
  <dict>
    <key>items</key>
    <array>
      <dict>
        <key>assets</key>
        <array>
          <dict>
            <key>kind</key>
            <string>software-package</string>
            <key>url</key>
            <string>{{ .Pkg }}</string>
          </dict>
          <dict>
            <key>kind</key>
            <string>display-image</string>
            <key>needs-shine</key>
            <true/>
            <key>url</key>
            <string>{{ .Icon }}</string>
          </dict>
        </array>
        <key>metadata</key>
        <dict>
          <key>bundle-identifier</key>
          <string>{{ .Identifier }}</string>
          <key>bundle-version</key>
          <string>{{ .Version }}</string>
          <key>kind</key>
          <string>software</string>
          <key>title</key>
          <string>{{ .Name }}</string>
        </dict>
      </dict>
    </array>
  </dict>
</plist>`

var defaultTemplate, _ = template.New("install-plist").Parse(plistTpl)

func NewInstallPlist(app *Item) ([]byte, error) {
	buf := bytes.NewBufferString("")
	err := defaultTemplate.Execute(buf, app)
	if err != nil {
		return nil, err
	}

	d, err := ioutil.ReadAll(buf)
	if err != nil {
		return nil, err
	}
	return d, nil
}
