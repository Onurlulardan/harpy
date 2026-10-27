package actorassertionv1

import (
	"encoding/base64"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func encodeClaims(claims Claims) (string, error) {
	if claims.Validate() != nil {
		return "", ErrContract
	}
	document, err := json.Marshal(claimsToWire(claims))
	if err != nil {
		return "", ErrContract
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumWireSize)
	if err != nil {
		return "", ErrContract
	}
	return base64.RawURLEncoding.EncodeToString(canonical), nil
}
