package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/iineva/ipa-server/pkg/common"
)

type param struct {
	publicURL string
	id        string
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

func DecodeListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return param{publicURL: publicURL(r)}, nil
}

func DecodeFindRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id := filepath.Base(r.URL.Path)
	if id == "" {
		return nil, ErrIdInvalid
	}
	_, err := uuid.Parse(id)
	if err != nil {
		// TODO: log error
		return nil, ErrIdInvalid
	}
	return param{publicURL: publicURL(r), id: id}, nil
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
