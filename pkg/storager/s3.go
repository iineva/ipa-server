package storager

import (
	"context"
	"io"
	"io/ioutil"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/iineva/ipa-server/pkg/storager/helper"
)

type s3Storager struct {
	endpoint string
	ak       string
	sk       string
	bucket   string
	domain   string
	client   *s3.Client
}

func NewS3Storager(endpoint, ak, sk, bucket, domain string) (Storager, error) {

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: u.String(),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(ak, sk, "")),
		config.WithEndpointResolver(customResolver),
	)
	if err != nil {
		return nil, err
	}

	return &s3Storager{
		endpoint: endpoint,
		ak:       ak,
		sk:       sk,
		bucket:   bucket,
		domain:   domain,
		client:   s3.NewFromConfig(cfg),
	}, nil
}

func (s *s3Storager) Save(name string, reader io.Reader) error {
	r := ioutil.NopCloser(reader) // avoid oss SDK to close reader
	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(name),
		Body:   r,
	}, s3.WithAPIOptions(
		v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
	))
	return err
}

func (s *s3Storager) OpenMetadata(name string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(name),
	})
	return out.Body, err
}

func (s *s3Storager) Delete(name string) error {
	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(name),
	})
	return err
}

func (s *s3Storager) Move(src, dest string) error {
	_, err := s.client.CopyObject(context.Background(), &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(s.bucket + "/" + src),
		Key:        aws.String(dest),
	})
	if err != nil {
		return err
	}
	return s.Delete(src)
}

func (s *s3Storager) PublicURL(publicURL, name string) (string, error) {
	return helper.UrlJoin(s.domain, name)
}
