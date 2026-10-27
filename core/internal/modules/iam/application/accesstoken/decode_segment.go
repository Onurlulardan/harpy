package accesstoken

import (
	"encoding/base64"
	"errors"
)

func base64URLStrict(segment string) ([]byte, error) {
	if segment == "" {
		return nil, errors.New("empty base64url segment")
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(segment)
	if err != nil || encodeBase64URL(decoded) != segment {
		return nil, errors.New("noncanonical base64url segment")
	}
	return decoded, nil
}
