package service

import (
	"errors"
	"testing"
)

const targetData = `<?xml version="1.0" encoding="utf-8"?>
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
            <string>https://file.example.com/ipa.ipa</string>
          </dict>
          <dict>
            <key>kind</key>
            <string>display-image</string>
            <key>needs-shine</key>
            <true/>
            <key>url</key>
            <string>https://file.example.com/icon.png</string>
          </dict>
        </array>
        <key>metadata</key>
        <dict>
          <key>bundle-identifier</key>
          <string>com.ineva.test</string>
          <key>bundle-version</key>
          <string>1.0</string>
          <key>kind</key>
          <string>software</string>
          <key>title</key>
          <string>Test</string>
        </dict>
      </dict>
    </array>
  </dict>
</plist>`

func TestNewInstallPlist(t *testing.T) {
	app := &Item{}
	app.Name = "Test"
	app.Version = "1.0"
	app.Identifier = "com.ineva.test"
	app.Icon = "https://file.example.com/icon.png"
	app.Pkg = "https://file.example.com/ipa.ipa"

	d, err := NewInstallPlist(app)
	if err != nil {
		t.Fatal(err)
	}

	if string(d) != targetData {
		t.Fatal(errors.New("created install plist not match"))
	}
}
