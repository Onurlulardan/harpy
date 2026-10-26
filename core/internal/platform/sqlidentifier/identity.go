package sqlidentifier

import (
	"crypto/sha256"
	"encoding/binary"
)

// Matches reports exact identity without exposing the opaque identifier.
func Matches(token Token, expected string) bool {
	return token.Validate() == nil && token.identifier == expected
}

// IsStatic reports whether token came from the static owner catalog minter.
func IsStatic(token Token) bool {
	return token.Validate() == nil && token.kind == tokenKindStatic
}

// Fingerprint returns a domain-separated identity digest without rendering SQL.
func Fingerprint(token Token) ([sha256.Size]byte, error) {
	if err := token.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	digest := sha256.New()
	_, _ = digest.Write([]byte("sqlidentifier_token_v1\x00"))
	_, _ = digest.Write([]byte{byte(token.kind)})
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(token.identifier)))
	_, _ = digest.Write(size[:])
	_, _ = digest.Write([]byte(token.identifier))
	var result [sha256.Size]byte
	copy(result[:], digest.Sum(nil))
	return result, nil
}
