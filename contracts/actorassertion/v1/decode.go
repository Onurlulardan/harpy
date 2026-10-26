package actorassertionv1

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func decodeClaims(segment string) (Claims, error) {
	document, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil || base64.RawURLEncoding.EncodeToString(document) != segment {
		return Claims{}, ErrContract
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumWireSize)
	if err != nil || !bytes.Equal(document, canonical) {
		return Claims{}, ErrContract
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var wire wireClaims
	if decoder.Decode(&wire) != nil {
		return Claims{}, ErrContract
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return Claims{}, ErrContract
	}
	return wire.typed()
}
