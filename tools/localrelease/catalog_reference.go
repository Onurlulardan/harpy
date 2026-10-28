package main

import (
	"crypto/sha256"
	"encoding/hex"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
)

const (
	localCatalogPackageID   = "10000000-0000-4000-8000-000000000031"
	localCatalogComponentID = "018f0000-0000-7000-8000-000000000032"
)

type catalogReference struct {
	PackageID      string
	Slug           string
	Version        string
	Manifest       []byte
	ManifestSHA256 string
	OwnerPayloads  []catalogpackagev1.OwnerPayload
}

func localCatalogReferences() ([]catalogReference, error) {
	payloadDigest := sha256.Sum256(nil)
	manifest, payload, err := localDemoDocuments(hex.EncodeToString(payloadDigest[:]))
	if err != nil {
		return nil, err
	}
	payloadDigest = sha256.Sum256(payload)
	manifest, payload, err = localDemoDocuments(hex.EncodeToString(payloadDigest[:]))
	if err != nil {
		return nil, err
	}
	manifestDigest := sha256.Sum256(manifest)
	return []catalogReference{{
		PackageID: localCatalogPackageID, Slug: "project", Version: "1.0.0",
		Manifest: manifest, ManifestSHA256: hex.EncodeToString(manifestDigest[:]),
		OwnerPayloads: []catalogpackagev1.OwnerPayload{{
			ComponentID: localCatalogComponentID, Owner: "catalog",
			Contract:      catalogpackagev1.ContractCatalog,
			ContentSHA256: hex.EncodeToString(payloadDigest[:]),
			Document:      payload,
		}},
	}}, nil
}
