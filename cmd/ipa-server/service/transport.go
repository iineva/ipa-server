package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-kit/kit/endpoint"
	"github.com/iineva/ipa-server/pkg/common"
)

type param struct {
	publicURL string
	id        string
}

type addParam struct {
	file multipart.File
	size int64
}

type data interface{}
type response struct {
	data
	Err string `json:"err"`
}

var (
	ErrIdInvalid = errors.New("id invalid")
)

func MakeListEndpoint(srv Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		p := request.(param)
		return srv.List(p.publicURL)
	}
}

func MakeFindEndpoint(srv Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		p := request.(param)
		return srv.Find(p.id, p.publicURL)
	}
}

func MakeAddEndpoint(srv Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		p := request.(addParam)
		defer p.file.Close()
		err := srv.Add(p.file, p.size)
		if err != nil {
			return nil, err
		}
		return map[string]string{"msg": "ok"}, nil
	}
}

func MakePlistEndpoint(srv Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		p := request.(param)

		d, err := srv.Plist(p.id, p.publicURL)
		if err != nil {
			return nil, err
		}
		return d, nil
	}
}

func DecodeListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// http://localhost/api/list
	return param{publicURL: publicURL(r)}, nil
}

func DecodeFindRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// http://localhost/api/info/{id}
	id := filepath.Base(r.URL.Path)
	if id == "" {
		return nil, ErrIdInvalid
	}

	if err := matchID(id); err != nil {
		return nil, ErrIdInvalid
	}
	return param{publicURL: publicURL(r), id: id}, nil
}

func DecodeAddRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// http://localhost/api/upload
	if r.Method != http.MethodPost {
		return nil, errors.New("404")
	}

	err := r.ParseMultipartForm(0)
	if err != nil {
		return nil, err
	}
	file, handler, err := r.FormFile("file")
	return addParam{file: file, size: handler.Size}, nil
}

func DecodePlistRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// http://localhost/plist/{id}.plist
	id := strings.TrimSuffix(filepath.Base(r.URL.Path), ".plist")
	if err := matchID(id); err != nil {
		return nil, ErrIdInvalid
	}

	return param{publicURL: publicURL(r), id: id}, nil
}

func EncodeJsonResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func EncodePlistResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	d := response.([]byte)
	n, err := io.Copy(w, bytes.NewBuffer(d))
	if err != nil {
		return err
	}
	if int64(len(d)) != n {
		return errors.New("wirte body len not match")
	}
	return nil
}

// auto check public url from frontend
func publicURL(ctx *http.Request) string {
	ref := ctx.Header.Get("referer")
	if ref != "" {
		u, _ := url.Parse(ref)
		return fmt.Sprintf("%v://%v", u.Scheme, u.Host)
	}

	xProto := ctx.Header.Get("x-forwarded-proto")
	host := ctx.Host
	return fmt.Sprintf("%v://%v", common.Def(xProto, "http"), host)
}

func matchID(id string) error {
	const idRegexp = `^[0-9a-zA-Z]{16,32}$`
	match, err := regexp.MatchString(idRegexp, id)
	if err != nil {
		return err
	}
	if !match {
		return ErrIdInvalid
	}
	return nil
}
