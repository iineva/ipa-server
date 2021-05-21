package plist

import (
	"errors"
	"os"
	"testing"
)

func TestPlist(t *testing.T) {
	fileName := "test_data/Info.plist"
	f, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	info, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	mustBe := map[string]string{
		"CFBundleExecutable":        "Test",
		"CFBundleIdentifier":        "com.ineva.test-rtmp.Test",
		"CFBundleDevelopmentRegion": "en",
	}
	for k, v := range mustBe {
		if info.GetString(k) != v {
			t.Fatal(errors.New("parse value error"))
		}
	}
}
