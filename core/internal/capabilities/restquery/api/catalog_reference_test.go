package api

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

const (
	catalogResourceID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	catalogVersionID  = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
)

func TestCatalogReferenceCanonicalRoundTrip(t *testing.T) {
	hash, _ := NewCatalogHash(validCatalogHash)
	reference, err := NewCatalogReference(
		"forms_engine_2", catalogResourceID, catalogVersionID, hash,
	)
	if err != nil || reference.Validate() != nil {
		t.Fatalf("valid catalog reference rejected: %v", err)
	}
	owner, ownerErr := reference.Owner()
	resourceID, resourceErr := reference.ResourceID()
	versionID, versionErr := reference.PublishedVersionID()
	actualHash, hashErr := reference.CatalogHash()
	actualHex, hexErr := actualHash.Hex()
	if ownerErr != nil || resourceErr != nil || versionErr != nil ||
		hashErr != nil || hexErr != nil {
		t.Fatal("valid catalog reference accessor failed")
	}
	if owner != "forms_engine_2" || resourceID.String() != catalogResourceID ||
		versionID.String() != catalogVersionID || actualHex != validCatalogHash {
		t.Fatal("catalog reference did not round trip")
	}
}

func TestCatalogReferenceRejectsInvalidOwnerUUIDOrHash(t *testing.T) {
	hash, _ := NewCatalogHash(validCatalogHash)
	invalidOwners := []string{"", "Forms", "2forms", "forms-api", "forms ", strings.Repeat("a", 64)}
	for _, owner := range invalidOwners {
		if _, err := NewCatalogReference(owner, catalogResourceID, catalogVersionID, hash); err == nil {
			t.Fatalf("invalid owner accepted: %q", owner)
		}
	}
	invalidIDs := []string{
		"", uuid.Nil.String(), strings.ToUpper(catalogResourceID),
		"{aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa}",
		"00000000-0000-4000-0000-000000000001",
	}
	for _, identifier := range invalidIDs {
		if _, err := NewCatalogReference("forms", identifier, catalogVersionID, hash); err == nil {
			t.Fatalf("invalid resource ID accepted: %q", identifier)
		}
		if _, err := NewCatalogReference("forms", catalogResourceID, identifier, hash); err == nil {
			t.Fatalf("invalid version ID accepted: %q", identifier)
		}
	}
	if _, err := NewCatalogReference("forms", catalogResourceID, catalogVersionID, CatalogHash{}); err == nil {
		t.Fatal("invalid catalog hash accepted")
	}
}

func TestCatalogReferenceZeroAccessorsFailClosed(t *testing.T) {
	var reference CatalogReference
	if reference.Validate() == nil {
		t.Fatal("zero catalog reference validated")
	}
	if _, err := reference.Owner(); err == nil {
		t.Fatal("zero reference exposed owner")
	}
	if _, err := reference.ResourceID(); err == nil {
		t.Fatal("zero reference exposed resource ID")
	}
	if _, err := reference.PublishedVersionID(); err == nil {
		t.Fatal("zero reference exposed version ID")
	}
	if _, err := reference.CatalogHash(); err == nil {
		t.Fatal("zero reference exposed catalog hash")
	}
}

func TestCatalogReferencePackageTamperFailsClosed(t *testing.T) {
	hash, _ := NewCatalogHash(validCatalogHash)
	valid, _ := NewCatalogReference("forms", catalogResourceID, catalogVersionID, hash)
	unparsed := valid
	unparsed.parsed = false
	invalidOwner := valid
	invalidOwner.owner = "PrivateOwner"
	missingResource := valid
	missingResource.resourceID = uuid.Nil
	missingVersion := valid
	missingVersion.publishedVersionID = uuid.Nil
	missingHash := valid
	missingHash.catalogHash = CatalogHash{}
	for _, reference := range []CatalogReference{
		unparsed, invalidOwner, missingResource, missingVersion, missingHash,
	} {
		if reference.Validate() == nil {
			t.Fatal("tampered catalog reference validated")
		}
	}
}
