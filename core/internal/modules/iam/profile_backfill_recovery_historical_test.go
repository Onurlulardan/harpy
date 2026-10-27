package iam

import (
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func TestProfileRecoveryInventoryAllowsEmptyHistoricalMappingSubset(t *testing.T) {
	fieldID := uuid.MustParse("12000000-0000-4000-8000-000000000001")
	historical := profileRecoveryRetiredSchema(t,
		"13000000-0000-4000-8000-000000000001", 1)
	source := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000002", 2, fieldID, false, 64)
	target := profilePublicationSchema(t,
		"13000000-0000-4000-8000-000000000003", 3, fieldID, false, 64)
	parent := profileRecoveryParent(t, source, target)
	inventory, err := classifyProfileRecoveryFixture(
		t, parent, source, target,
		map[uuid.UUID]profileapp.Schema{historical.ID(): historical},
		[]profilebackfillapp.SourceResource{profileRecoveryResource(
			"14000000-0000-4000-8000-000000000001", historical.ID(), []byte("{}"),
		)},
	)
	if err != nil {
		t.Fatal(err)
	}
	if inventory.InheritedTargetCount() != 0 || inventory.RemainingSourceCount() != 1 {
		t.Fatalf("inventory classification = %d/%d",
			inventory.InheritedTargetCount(), inventory.RemainingSourceCount())
	}
	mappings, err := profileBackfillMigrationMappings(parent.Mappings)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := profileapp.SelectMigrationMappingsForSource(
		historical, target, mappings,
	)
	if err != nil || len(selected) != 0 {
		t.Fatalf("historical mapping subset = %d/%v", len(selected), err)
	}
}
