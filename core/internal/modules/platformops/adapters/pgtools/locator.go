package pgtools

import (
	"crypto/sha256"
	"encoding/hex"
)

// LocatorSHA256 binds a database endpoint without retaining its password.
func LocatorSHA256(rawDSN string) (string, error) {
	sanitized, _, err := sanitizedDSN(rawDSN)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(sanitized))
	return hex.EncodeToString(digest[:]), nil
}
