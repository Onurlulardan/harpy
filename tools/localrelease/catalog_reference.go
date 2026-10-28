package main

import (
	"crypto/sha256"
	"encoding/hex"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
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
	payloadDocument := []byte(
		`{"component_id":"` + localCatalogComponentID +
			`","contract_version":"` + catalogpackagev1.ContractCatalog +
			`","module_slug":"project"}`,
	)
	payload, err := canonicaljsonv1.CanonicalizeObject(
		payloadDocument, catalogpackagev1.MaximumPayloadBytes,
	)
	if err != nil {
		return nil, err
	}
	payloadDigest := sha256.Sum256(payload)
	manifestDocument := []byte(
		`{"contract":"` + catalogpackagev1.ManifestContract +
			`","package_id":"` + localCatalogPackageID +
			`","slug":"project","version":"1.0.0"}`,
	)
	manifest, err := canonicaljsonv1.CanonicalizeObject(
		manifestDocument, catalogpackagev1.MaximumManifestBytes,
	)
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
