package presentation

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func EncodeDocument(
	document Document,
	catalog FieldCatalog,
) ([]byte, [sha256.Size]byte, error) {
	if err := Validate(document, catalog); err != nil {
		return nil, [sha256.Size]byte{}, err
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		return nil, [sha256.Size]byte{}, ErrInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(encoded, maximumPresentationBytes)
	if err != nil {
		return nil, [sha256.Size]byte{}, ErrInvalid
	}
	return canonical, sha256.Sum256(canonical), nil
}

func DecodeDocument(source []byte, catalog FieldCatalog) (Document, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(source, maximumPresentationBytes)
	if err != nil || !bytes.Equal(canonical, source) {
		return Document{}, ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	var document Document
	if decoder.Decode(&document) != nil {
		return Document{}, ErrInvalid
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF || Validate(document, catalog) != nil {
		return Document{}, ErrInvalid
	}
	return cloneDocument(document), nil
}

func CloneDocument(document Document) Document {
	return cloneDocument(document)
}

func cloneDocument(document Document) Document {
	encoded, err := json.Marshal(document)
	if err != nil {
		return Document{}
	}
	var result Document
	if json.Unmarshal(encoded, &result) != nil {
		return Document{}
	}
	return result
}
