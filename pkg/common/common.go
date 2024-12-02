package common

import "encoding/json"

// get args until arg is not empty
func Def(args ...string) string {
	for _, v := range args {
		if v != "" {
			return v
		}
	}
	return ""
}

// 结构体转 map
func ToMap(v interface{}) map[string]interface{} {
	b, err := json.Marshal(v)
	m := map[string]interface{}{}
	if err != nil {
		return m
	}
	_ = json.Unmarshal(b, &m)
	return m
}
