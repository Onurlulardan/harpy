package api

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/physicalid"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const validRelationID = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"

func mustRelationStorage(t *testing.T, relationID string) RelationStorage {
	t.Helper()
	identifier := uuid.MustParse(relationID)
	name, err := physicalid.Name("rel_", identifier)
	if err != nil {
		t.Fatal(err)
	}
	table, err := sqlidentifier.MintDynamic(name)
	if err != nil {
		t.Fatal(err)
	}
	storage, err := NewRelationStorage(relationID, table)
	if err != nil {
		t.Fatal(err)
	}
	return storage
}

func mustRelationTarget(t *testing.T) CatalogReference {
	t.Helper()
	hash, err := NewCatalogHash(validCatalogHash)
	if err != nil {
		t.Fatal(err)
	}
	target, err := NewCatalogReference("crm", catalogResourceID, catalogVersionID, hash)
	if err != nil {
		t.Fatal(err)
	}
	return target
}

func TestRelationDefinitionCanonicalRoundTrip(t *testing.T) {
	target := mustRelationTarget(t)
	definition, err := NewRelationDefinition(
		validRelationID, "müşteri", RelationCardinalityToMany,
		mustRelationStorage(t, validRelationID), target, true, 100,
	)
	if err != nil || definition.Validate() != nil {
		t.Fatalf("valid relation definition rejected: %v", err)
	}
	relationID, idErr := definition.RelationID()
	name, nameErr := definition.PublicName()
	cardinality, cardinalityErr := definition.Cardinality()
	actualTarget, targetErr := definition.Target()
	enabled, enabledErr := definition.IncludeEnabled()
	maximum, maximumErr := definition.MaximumIncludedRecords()
	if idErr != nil || nameErr != nil || cardinalityErr != nil || targetErr != nil ||
		enabledErr != nil || maximumErr != nil {
		t.Fatal("valid relation accessor failed")
	}
	if relationID.String() != validRelationID || name != "müşteri" ||
		cardinality != RelationCardinalityToMany || actualTarget != target ||
		!enabled || maximum != 100 {
		t.Fatal("relation definition did not round trip")
	}
	actualTarget.parsed = false
	again, err := definition.Target()
	if err != nil || again != target {
		t.Fatal("relation target was externally mutated")
	}
}

func TestRelationDefinitionRejectsInvalidIdentityNameOrTarget(t *testing.T) {
	target := mustRelationTarget(t)
	invalidIDs := []string{
		"", uuid.Nil.String(), strings.ToUpper(validRelationID),
		"{cccccccc-cccc-4ccc-8ccc-cccccccccccc}",
	}
	for _, identifier := range invalidIDs {
		if _, err := NewRelationDefinition(
			identifier, "customer", RelationCardinalityToOne,
			mustRelationStorage(t, validRelationID), target, false, 0,
		); err == nil {
			t.Fatalf("invalid relation ID accepted: %q", identifier)
		}
	}
	invalidNames := []string{"", "_included", "e\u0301", "nul\x00name", string([]byte{0xff})}
	for _, name := range invalidNames {
		if _, err := NewRelationDefinition(
			validRelationID, name, RelationCardinalityToOne,
			mustRelationStorage(t, validRelationID), target, false, 0,
		); err == nil {
			t.Fatalf("invalid public name accepted: %q", name)
		} else if name != "" && strings.Contains(err.Error(), name) {
			t.Fatalf("relation error leaked name: %v", err)
		}
	}
	if _, err := NewRelationDefinition(
		validRelationID, "customer", 0,
		mustRelationStorage(t, validRelationID), target, false, 0,
	); err == nil {
		t.Fatal("invalid cardinality accepted")
	}
	if _, err := NewRelationDefinition(
		validRelationID, "customer", RelationCardinalityToOne,
		mustRelationStorage(t, validRelationID),
		CatalogReference{}, false, 0,
	); err == nil {
		t.Fatal("invalid target accepted")
	}
}
