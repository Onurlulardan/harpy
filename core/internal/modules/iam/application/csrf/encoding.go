package csrf

import "encoding/base64"

func encode(value []byte) string { return base64.RawURLEncoding.EncodeToString(value) }

func decode(value string, size int) ([]byte, bool) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(decoded) != size || encode(decoded) != value {
		return nil, false
	}
	return decoded, true
}

func validKeyID(value string) bool {
	if len(value) < 16 || len(value) > 64 {
		return false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	return err == nil && len(decoded) >= 12 && len(decoded) <= 48 && encode(decoded) == value
}
