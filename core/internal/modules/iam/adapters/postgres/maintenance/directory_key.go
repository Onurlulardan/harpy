package maintenance

import (
	"crypto/sha256"
	"encoding/base64"
)

type DirectoryKey struct {
	ID       string
	Material [sha256.Size]byte
	Active   bool
}

func (key DirectoryKey) valid() bool {
	if !validDirectoryKeyID(key.ID) {
		return false
	}
	var aggregate byte
	for _, value := range key.Material {
		aggregate |= value
	}
	return aggregate != 0
}

func validDirectoryKeyID(value string) bool {
	if len(value) < 16 || len(value) > 64 {
		return false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	valid := err == nil && len(decoded) >= 12 && len(decoded) <= 48 &&
		base64.RawURLEncoding.EncodeToString(decoded) == value
	clear(decoded)
	return valid
}
