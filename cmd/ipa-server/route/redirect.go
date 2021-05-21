package route

import (
	"net/http"
)

func Redirect(m map[string]string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := m[r.URL.Path]
		if ok {
			r.URL.Path = p
		}
		next.ServeHTTP(w, r)
	})
}
