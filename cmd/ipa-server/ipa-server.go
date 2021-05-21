package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/iineva/ipa-server/cmd/ipa-server/route"
	"github.com/iineva/ipa-server/public"
)

func getEnv(key string, def ...string) string {
	v := os.Getenv(key)
	if v == "" && len(def) != 0 {
		v = def[0]
	}
	return v
}

func main() {

	host := fmt.Sprintf("%s:%s", getEnv("ADDRESS", "0.0.0.0"), getEnv("PORT", "8080"))

	r := route.New()

	// parser API
	http.HandleFunc("/api/list", r.HandlerList)
	http.HandleFunc("/api/info/", r.HandlerFind)

	// static files
	http.Handle("/", route.Redirect(map[string]string{
		"/key": "/key.html",
	}, http.FileServer(http.FS(public.FS))))

	log.Printf("SERVER LISTEN ON: http://%v", host)
	log.Fatal(http.ListenAndServe(host, nil))
}
