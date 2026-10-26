package custodymanifest

import (
	"encoding/base64"
	"fmt"
)

func decodeSecurityEpoch(value string) ([32]byte, error) {
	var result [32]byte
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) != len(result) {
		return result, fmt.Errorf("IAM custody security epoch is invalid")
	}
	copy(result[:], decoded)
	if result == [32]byte{} {
		return result, fmt.Errorf("IAM custody security epoch is invalid")
	}
	return result, nil
}
