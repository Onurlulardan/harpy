package domain

import (
	"encoding/base32"
	"errors"
)

var ErrInvalidObjectIdentity = errors.New("documents object identity is invalid")

func ContentObjectKey(tenantID, documentID, versionID ID) (string, error) {
	if !tenantID.Valid() || !documentID.UUIDv7() || !versionID.UUIDv7() {
		return "", ErrInvalidObjectIdentity
	}
	tenant := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(tenantID[:])
	return "tenants/" + lowerASCII(tenant) + "/documents/" + documentID.String() +
		"/versions/" + versionID.String() + "/content", nil
}

func ArchiveObjectKey(tenantID, archiveID ID) (string, error) {
	if !tenantID.Valid() || !archiveID.UUIDv7() {
		return "", ErrInvalidObjectIdentity
	}
	tenant := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(tenantID[:])
	return "tenants/" + lowerASCII(tenant) + "/document-archives/" +
		archiveID.String() + "/content", nil
}

func lowerASCII(value string) string {
	result := []byte(value)
	for index, character := range result {
		if character >= 'A' && character <= 'Z' {
			result[index] = character + ('a' - 'A')
		}
	}
	return string(result)
}
