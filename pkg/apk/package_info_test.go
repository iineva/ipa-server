package apk

import (
	"fmt"
	"testing"

	"github.com/shogo82148/androidbinary/apk"
)

func TestApkInfo(t *testing.T) {

	a, err := apk.OpenFile("./test.apk")
	if err != nil {
		t.Fatal(err)
	}

	for _, m := range a.Manifest().App.MetaData {
		name, _ := m.Name.String()
		value, _ := m.Value.String()
		fmt.Printf("%s: %s\n", name, value)
	}

}
