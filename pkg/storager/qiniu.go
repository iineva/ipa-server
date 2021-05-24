package storager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"

	"github.com/iineva/ipa-server/pkg/uuid"
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
// domain required: https://file.example.com
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

func (q *qiniuStorager) newUploadToken(keyToOverwrite string) string {
	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", q.bucket, keyToOverwrite),
	}
	return putPolicy.UploadToken(q.newMac())
}

func (q *qiniuStorager) newBucketManager() *storage.BucketManager {
	return storage.NewBucketManager(q.newMac(), q.config)
}

func (q *qiniuStorager) upload(name string, reader io.Reader) (*storage.PutRet, error) {
	resumeUploader := storage.NewResumeUploaderV2(q.config)
	ret := &storage.PutRet{}
	putExtra := storage.RputV2Extra{}
	err := resumeUploader.PutWithoutSize(context.Background(), ret, q.newUploadToken(name), name, reader, &putExtra)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (q *qiniuStorager) delete(name string) error {
	return q.newBucketManager().Delete(q.bucket, name)
}

func (q *qiniuStorager) copy(src string, dest string) error {
	return q.newBucketManager().Copy(q.bucket, src, q.bucket, dest, true)
}

func (q *qiniuStorager) Save(name string, reader io.Reader) error {
	_, err := q.upload(name, reader)
	return err
}

type runAfterReaderClose struct {
	cb     func() error
	reader io.ReadCloser
}

func (d *runAfterReaderClose) Close() error {
	if err := d.reader.Close(); err != nil {
		return err
	}
	return d.cb()
}

func (d *runAfterReaderClose) Read(p []byte) (int, error) {
	return d.reader.Read(p)
}

func (q *qiniuStorager) OpenMetadata(name string) (io.ReadCloser, error) {

	// copy to random file name to fix CDN cache
	// don not use refresh API, because it has rate limit
	targetName := fmt.Sprintf("temp-%v.json", uuid.NewString())
	err := q.copy(name, targetName)
	if err != nil {
		return nil, err
	}

	u := storage.MakePublicURL(q.domain.String(), targetName)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	return &runAfterReaderClose{reader: resp.Body, cb: func() error {
		return q.delete(targetName)
	}}, err
}

func (q *qiniuStorager) Delete(name string) error {
	return q.delete(name)
}

func (q *qiniuStorager) PublicURL(_, name string) (string, error) {
	d := q.domain
	d.Path = filepath.Join(d.Path, name)
	return d.String(), nil
}
