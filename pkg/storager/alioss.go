package storager

import (
	"io"
	"io/ioutil"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/iineva/ipa-server/pkg/storager/helper"
)

type aliossStorager struct {
	client *oss.Client
	bucket *oss.Bucket
	domain string
}

var _ Storager = (*aliossStorager)(nil)

// endpoint: https://help.aliyun.com/document_detail/31837.htm
func NewAliOssStorager(endpoint, accessKeyId, accessKeySecret, bucketName, domain string) (Storager, error) {
	client, err := oss.New(endpoint, accessKeyId, accessKeySecret, oss.Timeout(10, 120))
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}
	return &aliossStorager{client: client, bucket: bucket, domain: domain}, nil
}

func (a *aliossStorager) Save(name string, reader io.Reader) error {
	r := ioutil.NopCloser(reader) // avoid oss SDK to close reader
	return a.bucket.PutObject(name, r)
}

func (a *aliossStorager) OpenMetadata(name string) (io.ReadCloser, error) {
	return a.bucket.GetObject(name)
}

func (a *aliossStorager) Delete(name string) error {
	return a.bucket.DeleteObject(name)
}

func (a *aliossStorager) Move(src, dest string) error {
	_, err := a.bucket.CopyObject(src, dest)
	if err != nil {
		return err
	}
	return a.bucket.DeleteObject(src)
}

func (a *aliossStorager) PublicURL(publicURL, name string) (string, error) {
	return helper.UrlJoin(a.domain, name)
}
