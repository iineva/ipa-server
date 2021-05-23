package storager

import (
	"os"
	"testing"
)

func TestQiniuUpload(t *testing.T) {
	q, err := NewQiniuStorager("", "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}

	fileName := "../../public/img/default.png"
	name := "test.png"
	f, err := os.Open(fileName)
	defer f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if err := q.Save(name, f); err != nil {
		t.Fatal(err)
	}

	if err := q.Delete(name); err != nil {
		t.Fatal(err)
	}
}
