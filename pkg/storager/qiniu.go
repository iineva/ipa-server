package storager

import (
	"context"
	"errors"
	"io"
	"net/url"
	"path/filepath"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

type qiniuStorager struct {
	bucket    string
	accessKey string
	secretKey string
	domain    url.URL
	config    *storage.Config
}

var _ Storager = (*qiniuStorager)(nil)

var (
	ErrQiniuZoneCodeNotFound = errors.New("qiniu zone code not found")
)

// zone option: huadong:z0 huabei:z1 huanan:z2 northAmerica:na0 singapore:as0 fogCnEast1:fog-cn-east-1
// domain required: https://file.example.com/path/to/dir
func NewQiniuStorager(accessKey, secretKey, zone, bucket, domain string) (Storager, error) {
	config := &storage.Config{
		UseHTTPS:      true,
		UseCdnDomains: false,
	}
	// try get zone
	if zone != "" {
		z, ok := storage.GetRegionByID(storage.RegionID(zone))
		if !ok {
			return nil, ErrQiniuZoneCodeNotFound
		}
		config.Zone = &z
	}

	d, err := url.Parse(domain)
	if err != nil {
		return nil, err
	}

	return &qiniuStorager{
		bucket:    bucket,
		accessKey: accessKey,
		secretKey: secretKey,
		config:    config,
		domain:    *d,
	}, nil
}

func (q *qiniuStorager) newMac() *auth.Credentials {
	return qbox.NewMac(q.accessKey, q.secretKey)
}

func (q *qiniuStorager) newUploadToken() string {
	putPolicy := storage.PutPolicy{
		Scope: q.bucket,
	}
	return putPolicy.UploadToken(q.newMac())
}

func (q *qiniuStorager) upload(name string, reader io.Reader) (*storage.PutRet, error) {
	resumeUploader := storage.NewResumeUploaderV2(q.config)
	ret := &storage.PutRet{}
	putExtra := storage.RputV2Extra{}
	err := resumeUploader.PutWithoutSize(context.Background(), ret, q.newUploadToken(), name, reader, &putExtra)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (q *qiniuStorager) delete(name string) error {
	bucketManager := storage.NewBucketManager(q.newMac(), q.config)
	return bucketManager.Delete(q.bucket, name)
}

func (q *qiniuStorager) Save(name string, reader io.Reader) error {
	_, err := q.upload(name, reader)
	return err
}

func (q *qiniuStorager) Delete(name string) error {
	return q.delete(name)
}

func (q *qiniuStorager) PublicURL(_, name string) (string, error) {
	d := q.domain
	d.Path = filepath.Join(d.Path, name)
	return d.String(), nil
}
