package recordkernel

import (
	"encoding/hex"
	"strings"
)

func parseHexDigest(value string) ([32]byte, error) {
	if len(value) != 64 || strings.ToLower(value) != value {
		return [32]byte{}, ErrCatalogInvalid
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		return [32]byte{}, ErrCatalogInvalid
	}
	var result [32]byte
	copy(result[:], decoded)
	return result, nil
}
