package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"

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

func DecodeListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return param{publicURL: publicURL(r)}, nil
}

func DecodeFindRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id := filepath.Base(r.URL.Path)
	if id == "" {
		return nil, ErrIdInvalid
	}

	const idRegexp = `^[0-9a-zA-Z]{16,32}$`
	if match, err := regexp.MatchString(idRegexp, id); err != nil || !match {
		// TODO: log error
		return nil, ErrIdInvalid
	}
	return param{publicURL: publicURL(r), id: id}, nil
}

func DecodeAddRequest(_ context.Context, r *http.Request) (interface{}, error) {
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

func EncodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func publicURL(ctx *http.Request) string {
	ref := ctx.Header.Get("referer")
	if ref != "" {
		u, _ := url.Parse(ref)
		return fmt.Sprintf("%v://%v", u.Scheme, u.Host)
	}

	xProto := ctx.Header.Get("x-forwarded-proto")
	host := ctx.Header.Get("host")
	return fmt.Sprintf("%v://%v", common.Def(xProto, "http"), host)
}
