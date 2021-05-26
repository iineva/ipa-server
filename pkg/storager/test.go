package storager

import (
	"os"
	"testing"
)

func testStorager(s Storager, t *testing.T) {
	fileName := "../../public/img/default.png"
	name := "test.png"
	f, err := os.Open(fileName)
	defer f.Close()
	if err != nil {
		t.Fatal(err)
	}
	// save first
	if err := s.Save(name, f); err != nil {
		t.Fatal(err)
	}
	// overwite
	f2, err := os.Open(fileName)
	defer f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Save(name, f2); err != nil {
		t.Fatal(err)
	}
	// open metadata
	if reader, err := s.OpenMetadata(name); err != nil {
		reader.Close()
		t.Fatal(err)
	}
	// delete file
	if err := s.Delete(name); err != nil {
		t.Fatal(err)
	}
}
