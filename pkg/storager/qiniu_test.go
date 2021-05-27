package storager

import (
	"testing"
)

func TestQiniuUpload(t *testing.T) {
	zone := ""
	accessKeyId := "<yourAccessKeyId>"
	accessKeySecret := "<yourAccessKeySecret>"
	bucketName := "<yourBucketName>"
	domain := "<yourDomain>"

	q, err := NewQiniuStorager(zone, accessKeyId, accessKeySecret, bucketName, domain)
	if err != nil {
		t.Fatal(err)
	}

	testStorager(q, t)
}
