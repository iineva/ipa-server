package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/spf13/afero"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/iineva/ipa-server/cmd/ipasd/service"
	"github.com/iineva/ipa-server/pkg/httpfs"
	"github.com/iineva/ipa-server/pkg/storager"
	"github.com/iineva/ipa-server/pkg/uuid"
	"github.com/iineva/ipa-server/public"
)

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

	addr := flag.String("addr", "0.0.0.0", "bind addr")
	port := flag.String("port", "8080", "bind port")
	debug := flag.Bool("d", false, "enable debug logging")
	storageDir := flag.String("dir", "upload", "upload data storage dir")
	publicURL := flag.String("public-url", "", "server public url")
	metadataPath := flag.String("meta-path", "appList.json", "metadata storage path, use random secret path to keep your metadata safer")
	enabledDelete := flag.Bool("del", false, "delete app enabled")
	remoteCfg := flag.String("remote", "", "remote storager config, s3://ENDPOINT:AK:SK:BUCKET, alioss://ENDPOINT:AK:SK:BUCKET, qiniu://[ZONE]:AK:SK:BUCKET")
	remoteURL := flag.String("remote-url", "", "remote storager public url, https://cdn.example.com")

	flag.Usage = usage
	flag.Parse()

	serve := http.NewServeMux()

	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.TimestampFormat(time.Now, "2006-01-02 15:04:05.000"), "caller", log.DefaultCaller)

	var store storager.Storager
	if *remoteCfg != "" && *remoteURL != "" {
		r := strings.Split(*remoteCfg, "://")
		if len(r) != 2 {
			usage()
			os.Exit(0)
		}
		args := strings.Split(r[1], ":")
		if len(args) != 4 {
			usage()
			os.Exit(0)
		}

		switch r[0] {
		case "s3":
			logger.Log("msg", "used s3 storager")
			s, err := storager.NewS3Storager(args[0], args[1], args[2], args[3], *remoteURL)
			if err != nil {
				panic(err)
			}
			store = s
		case "alioss":
			logger.Log("msg", "used alioss storager")
			s, err := storager.NewAliOssStorager(args[0], args[1], args[2], args[3], *remoteURL)
			if err != nil {
				panic(err)
			}
			store = s
		case "qiniu":
			logger.Log("msg", "used qiniu storager")
			s, err := storager.NewQiniuStorager(args[0], args[1], args[2], args[3], *remoteURL)
			if err != nil {
				panic(err)
			}
			store = s
		}
	} else {
		logger.Log("msg", "used os file storager")
		store = storager.NewOsFileStorager(*storageDir)
	}

	srv := service.New(store, *publicURL, *metadataPath)
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
	deleteHandler := httptransport.NewServer(
		service.LoggingMiddleware(logger, "/api/delete", *debug)(service.MakeDeleteEndpoint(srv, *enabledDelete)),
		service.DecodeDeleteRequest,
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
	serve.Handle("/api/delete", deleteHandler)
	serve.Handle("/plist/", plistHandler)

	// static files
	uploadFS := afero.NewBasePathFs(afero.NewOsFs(), *storageDir)
	staticFS := httpfs.New(
		http.FS(public.FS),
		httpfs.NewAferoFS(uploadFS),
	)
	serve.Handle("/", redirect(map[string]string{
		"/key": "/key.html",
		// random path to block local metadata
		fmt.Sprintf("/%s", *metadataPath): fmt.Sprintf("/%s", uuid.NewString()),
	}, http.FileServer(staticFS)))

	host := fmt.Sprintf("%s:%s", *addr, *port)
	logger.Log("msg", fmt.Sprintf("SERVER LISTEN ON: http://%v", host))
	logger.Log("msg", http.ListenAndServe(host, serve))
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: ipa-server [options]
Options:
`)
	flag.PrintDefaults()
}
