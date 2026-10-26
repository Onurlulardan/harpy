package accesstoken

import "encoding/base64"

func encodeBase64URL(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}

func decodeBase64URL(value string, expectedBytes int) ([]byte, bool) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(decoded) != expectedBytes || encodeBase64URL(decoded) != value {
		return nil, false
	}
	return decoded, true
}

func validKeyID(value string) bool {
	if len(value) < 16 || len(value) > 64 {
		return false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	return err == nil && len(decoded) >= 12 && len(decoded) <= 48 &&
		encodeBase64URL(decoded) == value
}
