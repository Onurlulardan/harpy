package iam

import (
	"crypto/sha256"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func systemTransferSHA256(value []byte) [32]byte { return sha256.Sum256(value) }

func systemTransferKeySHA256(value string) ([32]byte, error) {
	if !systemTransferKeyPattern.MatchString(value) {
		return [32]byte{}, ErrSystemTransferInvalid
	}
	return sha256.Sum256([]byte(value)), nil
}

func systemTransferRequestSHA256(value any) ([32]byte, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return [32]byte{}, ErrSystemTransferInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 2<<20)
	if err != nil {
		return [32]byte{}, ErrSystemTransferInvalid
	}
	return sha256.Sum256(canonical), nil
}
