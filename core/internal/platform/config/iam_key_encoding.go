package config

import (
	"encoding/base64"
	"time"
)

func decodeIAMKeyMaterial(value string) ([32]byte, bool) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(decoded) != 32 ||
		base64.RawURLEncoding.EncodeToString(decoded) != value {
		clear(decoded)
		return [32]byte{}, false
	}
	var material [32]byte
	copy(material[:], decoded)
	clear(decoded)
	return material, true
}

func validIAMAccessKeyID(value string) bool {
	if len(value) < 16 || len(value) > 64 {
		return false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	valid := err == nil && len(decoded) >= 12 && len(decoded) <= 48 &&
		base64.RawURLEncoding.EncodeToString(decoded) == value
	clear(decoded)
	return valid
}

func validIAMHMACKeyID(value string) bool {
	return validIAMAccessKeyID(value)
}

func parseIAMTimestamp(value string) (time.Time, bool) {
	parsed, err := time.Parse("2006-01-02T15:04:05Z", value)
	return parsed, err == nil && parsed.Format("2006-01-02T15:04:05Z") == value
}
