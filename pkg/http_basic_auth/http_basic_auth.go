package http_basic_auth

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

var ErrUnauthorized = errors.New("Unauthorized")

// Returns a hash of a given slice.
func toHashSlice(s []byte) []byte {
	hash := sha256.Sum256(s)
	return hash[:]
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ([]byte("Aladdin"), []byte("open sesame"), true).
func parseBasicAuth(auth string) (username, password []byte, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}

	s := bytes.IndexByte(c, ':')
	if s < 0 {
		return
	}
	return c[:s], c[s+1:], true
}

func HandleBasicAuth(requiredUser, requiredPassword, realm string, r *http.Request) error {
	requiredUserBytes := toHashSlice([]byte(requiredUser))
	requiredPasswordBytes := toHashSlice([]byte(requiredPassword))

	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ErrUnauthorized
	}

	givenUser, givenPassword, ok := parseBasicAuth(auth)
	if !ok {
		return ErrUnauthorized
	}

	givenUserBytes := toHashSlice(givenUser)
	givenPasswordBytes := toHashSlice(givenPassword)

	if subtle.ConstantTimeCompare(givenUserBytes, requiredUserBytes) == 0 ||
		subtle.ConstantTimeCompare(givenPasswordBytes, requiredPasswordBytes) == 0 {
		return ErrUnauthorized
	}

	return nil
}
