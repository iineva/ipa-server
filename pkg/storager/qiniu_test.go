package storager

import (
	"testing"
)

func TestQiniuUpload(t *testing.T) {
	q, err := NewQiniuStorager("", "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}

	testStorager(q, t)
}
