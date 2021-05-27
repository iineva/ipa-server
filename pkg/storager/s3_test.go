package storager

import (
	"testing"
)

func TestS3(t *testing.T) {

	endpoint := "oss-cn-shenzhen.aliyuncs.com"
	accessKeyId := "<yourAccessKeyId>"
	accessKeySecret := "<yourAccessKeySecret>"
	bucketName := "<yourBucketName>"
	domain := "<yourDomain>"

	a, err := NewS3Storager(endpoint, accessKeyId, accessKeySecret, bucketName, domain)
	if err != nil {
		t.Fatal(err)
	}

	testStorager(a, t)
}
