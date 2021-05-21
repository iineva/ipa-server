package plist

import (
	"io"

	"howett.net/plist"
)

type Plist map[string]interface{}

func Parse(r io.ReadSeeker) (Plist, error) {
	decoder := plist.NewDecoder(r)
	p := Plist{}
	err := decoder.Decode(&p)
	if err != nil {
		return p, err
	}
	return p, nil
}

func (p *Plist) GetString(k string) string {
	if v, ok := (*p)[k]; ok {
		if value, ok := v.(string); ok {
			return value
		}
	}
	return ""
}
