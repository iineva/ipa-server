package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/spf13/afero"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/iineva/ipa-server/cmd/ipa-server/service"
	"github.com/iineva/ipa-server/pkg/httpfs"
	"github.com/iineva/ipa-server/pkg/storager"
	"github.com/iineva/ipa-server/public"
)

func getEnv(key string, def ...string) string {
	v := os.Getenv(key)
	if v == "" && len(def) != 0 {
		v = def[0]
	}
	return v
}

func redirect(m map[string]string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := m[r.URL.Path]
		if ok {
			r.URL.Path = p
		}
		next.ServeHTTP(w, r)
	})
}

func main() {

	debug := flag.Bool("d", false, "enable debug logging")
	storageDir := flag.String("dir", "data", "data storage dir")
	flag.Usage = usage
	flag.Parse()

	host := fmt.Sprintf("%s:%s", getEnv("ADDRESS", "0.0.0.0"), getEnv("PORT", "8080"))

	serve := http.NewServeMux()
	// r := route.New()

	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.TimestampFormat(time.Now, "2006-01-02 15:04:05.000"), "caller", log.DefaultCaller)

	srv := service.New(storager.NewOsFileStorager(*storageDir))
	listHandler := httptransport.NewServer(
		service.LoggingMiddleware(logger, "/api/list", *debug)(service.MakeListEndpoint(srv)),
		service.DecodeListRequest,
		service.EncodeResponse,
	)
	findHandler := httptransport.NewServer(
		service.LoggingMiddleware(logger, "/api/info", *debug)(service.MakeFindEndpoint(srv)),
		service.DecodeFindRequest,
		service.EncodeResponse,
	)
	addHandler := httptransport.NewServer(
		service.LoggingMiddleware(logger, "/api/upload", *debug)(service.MakeAddEndpoint(srv)),
		service.DecodeAddRequest,
		service.EncodeResponse,
	)

	// parser API
	serve.Handle("/api/list", listHandler)
	serve.Handle("/api/info/", findHandler)
	serve.Handle("/api/upload", addHandler)

	// static files
	uploadFS := afero.NewBasePathFs(afero.NewOsFs(), *storageDir)
	staticFS := httpfs.New(
		http.FS(public.FS),
		http.FS(httpfs.NewAferoFS(uploadFS)),
	)
	serve.Handle("/", redirect(map[string]string{
		"/key": "/key.html",
	}, http.FileServer(staticFS)))

	logger.Log("msg", fmt.Sprintf("SERVER LISTEN ON: http://%v", host))
	logger.Log("msg", http.ListenAndServe(host, serve))
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: ipa-server [options]
Options:
`)
	flag.PrintDefaults()
}
