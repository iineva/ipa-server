package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-kit/kit/endpoint"
	"github.com/iineva/ipa-server/pkg/common"
	pkgMultipart "github.com/iineva/ipa-server/pkg/multipart"
	"github.com/iineva/ipa-server/pkg/seekbuf"
)

type param struct {
	publicURL string
	id        string
}

type delParam struct {
	publicURL string
	id        string
	get       bool // get if delete enabled
}

type addParam struct {
	file *pkgMultipart.FormFile
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
		buf, err := seekbuf.Open(p.file, seekbuf.FileMode)
		if err != nil {
			return nil, err
		}
		defer buf.Close()

		t := FileType(p.file.FileName())
		if t == AppInfoTypeUnknown {
			return nil, fmt.Errorf("do not support %s file", path.Ext(p.file.FileName()))
		}

		if err := srv.Add(buf, t); err != nil {
			return nil, err
		}
		return map[string]string{"msg": "ok"}, nil
	}
}

func MakeDeleteEndpoint(srv Service, enabledDelete bool) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		p := request.(delParam)
		if p.get {
			// check is delete enabled
			return map[string]interface{}{"delete": enabledDelete}, nil
		}

		if !enabledDelete {
			return nil, errors.New("no permission to delete")
		}

		err := srv.Delete(p.id)
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

	if err := tryMatchID(id); err != nil {
		return nil, ErrIdInvalid
	}
	return param{publicURL: publicURL(r), id: id}, nil
}

func DecodeAddRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// http://localhost/api/upload
	if r.Method != http.MethodPost {
		return nil, errors.New("404")
	}

	m := pkgMultipart.New(r)
	f, err := m.GetFormFile("file")
	if err != nil {
		return nil, err
	}

	return addParam{file: f}, nil
}

func DecodeDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// http://localhost/api/delete

	if r.Method == http.MethodGet {
		return delParam{get: true}, nil
	}

	p := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		return nil, err
	}

	id := p["id"]
	if err := tryMatchID(id); err != nil {
		return nil, err
	}

	return delParam{id: id, get: false}, nil
}

func DecodePlistRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// http://localhost/plist/{id}.plist
	id := strings.TrimSuffix(filepath.Base(r.URL.Path), ".plist")
	if err := tryMatchID(id); err != nil {
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

func tryMatchID(id string) error {
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
