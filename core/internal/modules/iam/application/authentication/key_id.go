package authentication

import (
	"encoding/base64"

	kmsv1 "github.com/Onurlulardan/harpy/contracts/kms/v1"
)

const (
	minimumHMACKeyIDBytes = 16
	maximumHMACKeyIDBytes = 64
)

func validHMACKeyID(value string) bool {
	if len(value) < minimumHMACKeyIDBytes || len(value) > maximumHMACKeyIDBytes {
		return false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	valid := err == nil && len(decoded) >= 12 && len(decoded) <= 48 &&
		base64.RawURLEncoding.EncodeToString(decoded) == value
	clear(decoded)
	return valid
}

func validKMSKeyID(value string) bool {
	_, err := kmsv1.ParseGoogleCloudKeyVersion(value)
	return err == nil
}
