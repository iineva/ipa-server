package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/spf13/afero"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/iineva/ipa-server/cmd/ipa-server/service"
	"github.com/iineva/ipa-server/pkg/httpfs"
	"github.com/iineva/ipa-server/pkg/ipa"
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
	storageDir := flag.String("dir", "upload", "upload data storage dir")
	publicURL := flag.String("public-url", "", "public url")
	flag.Usage = usage
	flag.Parse()

	host := fmt.Sprintf("%s:%s", getEnv("ADDRESS", "0.0.0.0"), getEnv("PORT", "8080"))

	serve := http.NewServeMux()
	// r := route.New()

	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.TimestampFormat(time.Now, "2006-01-02 15:04:05.000"), "caller", log.DefaultCaller)

	srv := service.New(storager.NewOsFileStorager(*storageDir), *publicURL)
	listHandler := httptransport.NewServer(
		service.LoggingMiddleware(logger, "/api/list", *debug)(service.MakeListEndpoint(srv)),
		service.DecodeListRequest,
		service.EncodeJsonResponse,
	)
	findHandler := httptransport.NewServer(
		service.LoggingMiddleware(logger, "/api/info", *debug)(service.MakeFindEndpoint(srv)),
		service.DecodeFindRequest,
		service.EncodeJsonResponse,
	)
	addHandler := httptransport.NewServer(
		service.LoggingMiddleware(logger, "/api/upload", *debug)(service.MakeAddEndpoint(srv)),
		service.DecodeAddRequest,
		service.EncodeJsonResponse,
	)
	plistHandler := httptransport.NewServer(
		service.LoggingMiddleware(logger, "/plist", *debug)(service.MakePlistEndpoint(srv)),
		service.DecodePlistRequest,
		service.EncodePlistResponse,
	)

	// parser API
	serve.Handle("/api/list", listHandler)
	serve.Handle("/api/info/", findHandler)
	serve.Handle("/api/upload", addHandler)
	serve.Handle("/plist/", plistHandler)

	// static files
	uploadFS := afero.NewBasePathFs(afero.NewOsFs(), *storageDir)
	staticFS := httpfs.New(
		http.FS(public.FS),
		httpfs.NewAferoFS(uploadFS),
	)
	serve.Handle("/", redirect(map[string]string{
		"/key": "/key.html",
	}, http.FileServer(staticFS)))

	// try migrate old version data
	err := tryMigrateOldData(uploadFS, srv)
	if err != nil {
		logger.Log(
			"msg", "migrate old version data err",
			"err", err.Error(),
		)
	}

	logger.Log("msg", fmt.Sprintf("SERVER LISTEN ON: http://%v", host))
	logger.Log("msg", http.ListenAndServe(host, serve))
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: ipa-server [options]
Options:
`)
	flag.PrintDefaults()
}

func tryMigrateOldData(uploadFS afero.Fs, srv service.Service) error {
	f, err := uploadFS.Open("appList.json")
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	list := ipa.AppList{}
	if err := json.Unmarshal(b, &list); err != nil {
		return err
	}
	if err := srv.MigrateOldData(list); err != nil {
		return err
	}

	// TODO: delete old file

	return nil
}
