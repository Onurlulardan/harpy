package licenseauthoritysource

import (
	"encoding/base64"
	"errors"
	"strings"
)

var ErrInvalidConfiguration = errors.New("license source configuration is invalid")

func parseCredential(document []byte) (string, error) {
	value := string(document)
	value = strings.TrimSuffix(value, "\n")
	value = strings.TrimSuffix(value, "\r")
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(decoded) != 32 ||
		base64.RawURLEncoding.EncodeToString(decoded) != value {
		return "", ErrInvalidConfiguration
	}
	return value, nil
}
