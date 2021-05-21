package public

import "embed"

//go:embed app
//go:embed css
//go:embed img
//go:embed js
//go:embed *.html
var FS embed.FS
