
// create plist content
const createPlistBody = (opt = {}) => `<?xml version="1.0" encoding="utf-8"?>
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
            <string>${opt.ipa}</string>
          </dict>
          <dict>
            <key>kind</key>
            <string>display-image</string>
            <key>needs-shine</key>
            <true/>
            <key>url</key>
            <string>${opt.icon}</string>
          </dict>
        </array>
        <key>metadata</key>
        <dict>
          <key>bundle-identifier</key>
          <string>${opt.identifier}</string>
          <key>bundle-version</key>
          <string>${opt.version}</string>
          <key>kind</key>
          <string>software</string>
          <key>title</key>
          <string>${opt.name}</string>
        </dict>
      </dict>
    </array>
  </dict>
</plist>`

module.exports = {
  createPlistBody
}
