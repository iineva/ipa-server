package route

import (
	"encoding/json"
	"net/http"
	"path/filepath"
)

func (ro *Route) HandlerList(w http.ResponseWriter, r *http.Request) {
	list := ro.ipa.List(PublicURL(r))
	b, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func (ro *Route) HandlerFind(w http.ResponseWriter, r *http.Request) {
	id := filepath.Base(r.URL.Path)
	list := ro.ipa.Find(id, PublicURL(r))
	b, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
