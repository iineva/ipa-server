package storager

import (
	"testing"
)

func TestAliOss(t *testing.T) {

	endpoint := "oss-cn-shenzhen.aliyuncs.com"
	accessKeyId := "<yourAccessKeyId>"
	accessKeySecret := "<yourAccessKeySecret>"
	bucketName := "<yourBucketName>"
	domain := "<yourDomain>"

	a, err := NewAliOssStorager(endpoint, accessKeyId, accessKeySecret, bucketName, domain)
	if err != nil {
		t.Fatal(err)
	}

	testStorager(a, t)
}
