package route

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/iineva/ipa-server/pkg/common"
)

func PublicURL(ctx *http.Request) string {
	ref := ctx.Header.Get("referer")
	if ref != "" {
		u, _ := url.Parse(ref)
		return fmt.Sprintf("%v://%v", u.Scheme, u.Host)
	}

	xProto := ctx.Header.Get("x-forwarded-proto")
	host := ctx.Header.Get("host")
	return fmt.Sprintf("%v://%v", common.Def(xProto, "http"), host)
}
