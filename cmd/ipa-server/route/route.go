package route

import (
	"time"

	"github.com/iineva/ipa-server/pkg/ipa"
)

type Route struct {
	ipa ipa.IPAList
}

func New() *Route {
	return &Route{
		ipa: ipa.IPAList{
			&ipa.IPA{
				ID:         "xxxxx2",
				Name:       "Test",
				Version:    "1.0",
				Identifier: "com.ineva.test",
				Build:      "129",
				Channel:    "",
				Date:       time.Now(),
				Size:       1204 * 1024 * 10,
				NoneIcon:   true,
			},
			&ipa.IPA{
				ID:         "xxxxx1",
				Name:       "Test",
				Version:    "1.0",
				Identifier: "com.ineva.test",
				Build:      "128",
				Channel:    "",
				Date:       time.Now(),
				Size:       1204 * 1024,
				NoneIcon:   true,
			},
		},
	}
}
