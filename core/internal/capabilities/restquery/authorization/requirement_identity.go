package authorization

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/google/uuid"
)

const requirementIdentityVersion = "restquery_authorization_requirement_v1"

type requirementIdentityDocument struct {
	Version            string `json:"version"`
	Kind               string `json:"kind"`
	Owner              string `json:"owner"`
	ResourceID         string `json:"resource_id"`
	PublishedVersionID string `json:"published_version_id"`
	CatalogSHA256      string `json:"catalog_sha256"`
	ResourcePath       string `json:"resource_path"`
	RequestPath        string `json:"request_path"`
	FieldID            string `json:"field_id"`
	RelationID         string `json:"relation_id"`
}

func (requirement Requirement) identitySHA256() ([sha256.Size]byte, error) {
	document, err := requirement.identityDocument()
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	canonical, err := attestation.MarshalCanonical(document)
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	return sha256.Sum256(canonical), nil
}

func (requirement Requirement) identityDocument() (requirementIdentityDocument, error) {
	owner, ownerErr := requirement.catalog.Owner()
	resourceID, resourceErr := requirement.catalog.ResourceID()
	versionID, versionErr := requirement.catalog.PublishedVersionID()
	catalogHash, hashErr := requirement.catalog.CatalogHash()
	hash, hexErr := catalogHash.Hex()
	resourcePath, resourcePathErr := requirement.resourcePath.Value()
	requestPath, requestPathErr := requirement.requestPath.Value()
	if ownerErr != nil || resourceErr != nil || versionErr != nil || hashErr != nil ||
		hexErr != nil || resourcePathErr != nil || requestPathErr != nil ||
		requirement.kind.Validate() != nil {
		return requirementIdentityDocument{}, errInvalidRequirementIdentity
	}
	document := requirementIdentityDocument{
		Version: requirementIdentityVersion, Kind: requirement.kind.String(), Owner: owner,
		ResourceID: resourceID.String(), PublishedVersionID: versionID.String(),
		CatalogSHA256: hash, ResourcePath: resourcePath, RequestPath: requestPath,
	}
	if requirement.fieldID != uuid.Nil {
		document.FieldID = requirement.fieldID.String()
	}
	if requirement.relationID != uuid.Nil {
		document.RelationID = requirement.relationID.String()
	}
	return document, nil
}
