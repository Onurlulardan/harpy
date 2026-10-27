package dashboard

import (
	"bytes"
	"crypto/sha256"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

const MaximumDocumentBytes = 64 << 10

type Document struct {
	canonical []byte
	digest    [32]byte
}

func NewDocument(value []byte) (Document, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(value, MaximumDocumentBytes)
	if err != nil {
		return Document{}, ErrInvalid
	}
	return Document{canonical: bytes.Clone(canonical), digest: sha256.Sum256(canonical)}, nil
}

func (value Document) Valid() bool {
	other, err := NewDocument(value.canonical)
	return err == nil && other.digest == value.digest
}

func (value Document) CanonicalJSON() []byte { return bytes.Clone(value.canonical) }
func (value Document) SHA256() [32]byte      { return value.digest }
