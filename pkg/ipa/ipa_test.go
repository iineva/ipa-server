package ipa

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/iineva/ipa-server/pkg/seekbuf"
)

func TestReadPlistInfo(t *testing.T) {

	printMemUsage()

	fileName := "test_data/ipa.ipa"
	// fileName := "/Users/steven/Downloads/TikTok (18.5.0) Unicorn v4.9.ipa"
	f, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	buf, err := seekbuf.Open(f, seekbuf.MemoryMode)
	info, err := ReadPlist(buf, fi.Size())

	if err != nil {
		t.Fatal(err)
	}
	if info == nil {
		t.Fatal(errors.New("parse error"))
	}
	buf.Close()
	f.Close()
	printMemUsage()

	if info.Name != "Test" {
		t.Fatal(errors.New("parse error"))
	}
	// log.Printf("%+v", info)
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
