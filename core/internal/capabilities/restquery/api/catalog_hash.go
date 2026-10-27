package api

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

var errInvalidCatalogHash = errors.New("REST Query catalog hash is invalid")

type CatalogHash struct {
	digest [sha256.Size]byte
	parsed bool
}

func NewCatalogHash(value string) (CatalogHash, error) {
	if len(value) != sha256.Size*2 {
		return CatalogHash{}, errInvalidCatalogHash
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || hex.EncodeToString(decoded) != value {
		return CatalogHash{}, errInvalidCatalogHash
	}
	var digest [sha256.Size]byte
	copy(digest[:], decoded)
	if digest == [sha256.Size]byte{} {
		return CatalogHash{}, errInvalidCatalogHash
	}
	return CatalogHash{digest: digest, parsed: true}, nil
}

func (hash CatalogHash) Validate() error {
	if !hash.parsed || hash.digest == [sha256.Size]byte{} {
		return errInvalidCatalogHash
	}
	return nil
}

func (hash CatalogHash) Hex() (string, error) {
	if err := hash.Validate(); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.digest[:]), nil
}

func (hash CatalogHash) Bytes() ([]byte, error) {
	if err := hash.Validate(); err != nil {
		return nil, err
	}
	return append([]byte(nil), hash.digest[:]...), nil
}

func (CatalogHash) String() string   { return queryValueRedaction }
func (CatalogHash) GoString() string { return queryValueRedaction }
func (CatalogHash) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (CatalogHash) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
