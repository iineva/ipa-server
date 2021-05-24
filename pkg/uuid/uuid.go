package uuid

import "github.com/lithammer/shortuuid"

func NewString() string {
	return shortuuid.New()
}
