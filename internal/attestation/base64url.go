package attestation

import (
	"encoding/base64"
	"fmt"
)

func decodeBase64URLExact(value string, expectedLength int) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(decoded) != expectedLength ||
		base64.RawURLEncoding.EncodeToString(decoded) != value {
		return nil, fmt.Errorf("base64url value is not canonical")
	}
	return decoded, nil
}
