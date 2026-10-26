package api

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

var (
	errInvalidCatalogReference = errors.New("REST Query catalog reference is invalid")
	catalogOwnerPattern        = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
)

type CatalogReference struct {
	owner              string
	resourceID         uuid.UUID
	publishedVersionID uuid.UUID
	catalogHash        CatalogHash
	parsed             bool
}

func NewCatalogReference(
	owner string,
	resourceID string,
	publishedVersionID string,
	catalogHash CatalogHash,
) (CatalogReference, error) {
	resource, resourceOK := parseCatalogUUID(resourceID)
	publishedVersion, publishedVersionOK := parseCatalogUUID(publishedVersionID)
	if !catalogOwnerPattern.MatchString(owner) || !resourceOK || !publishedVersionOK ||
		catalogHash.Validate() != nil {
		return CatalogReference{}, errInvalidCatalogReference
	}
	return CatalogReference{
		owner: owner, resourceID: resource, publishedVersionID: publishedVersion,
		catalogHash: catalogHash, parsed: true,
	}, nil
}

func (reference CatalogReference) Validate() error {
	if !reference.parsed || !catalogOwnerPattern.MatchString(reference.owner) ||
		reference.resourceID == uuid.Nil || reference.publishedVersionID == uuid.Nil ||
		reference.catalogHash.Validate() != nil {
		return errInvalidCatalogReference
	}
	return nil
}

func (reference CatalogReference) Owner() (string, error) {
	if err := reference.Validate(); err != nil {
		return "", err
	}
	return reference.owner, nil
}

func (reference CatalogReference) ResourceID() (uuid.UUID, error) {
	if err := reference.Validate(); err != nil {
		return uuid.Nil, err
	}
	return reference.resourceID, nil
}

func (reference CatalogReference) PublishedVersionID() (uuid.UUID, error) {
	if err := reference.Validate(); err != nil {
		return uuid.Nil, err
	}
	return reference.publishedVersionID, nil
}

func (reference CatalogReference) CatalogHash() (CatalogHash, error) {
	if err := reference.Validate(); err != nil {
		return CatalogHash{}, err
	}
	return reference.catalogHash, nil
}

func (CatalogReference) String() string   { return queryValueRedaction }
func (CatalogReference) GoString() string { return queryValueRedaction }
func (CatalogReference) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (CatalogReference) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func parseCatalogUUID(value string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(value)
	return parsed, err == nil && parsed != uuid.Nil && parsed.String() == value &&
		parsed.Variant() == uuid.RFC4122
}
