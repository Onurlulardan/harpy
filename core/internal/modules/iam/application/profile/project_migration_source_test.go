package profile

import (
	"testing"

	"github.com/google/uuid"
)

func TestProjectMigrationSourceUsesRetiredCompatibilityValue(t *testing.T) {
	sourceField := mustField(t, baseFieldInput(fieldOneID, "alias", LogicalString))
	source := mustSchema(t, schemaOneID, 1, sourceField)
	retiredInput := baseFieldInput(fieldOneID, "alias", LogicalString)
	retiredInput.State = FieldRetired
	retired := mustField(t, retiredInput)
	targetID := uuid.MustParse("55555555-5555-4555-8555-555555555555")
	targetInput := baseFieldInput(targetID, "alias_v2", LogicalString)
	targetInput.Normalization = NormalizationEmailV1
	targetField := mustField(t, targetInput)
	target := mustSchema(t, schemaTwoID, 2, retired, targetField)
	mappings := []MigrationMapping{{
		SourceFieldID: fieldOneID, TargetFieldID: targetID,
		Converter: MigrationNormalize,
	}}
	migrated, err := MigrateDocument(
		source, target, mappings,
		[]byte(`{"11111111-1111-4111-8111-111111111111":"ADA@EXAMPLE.COM"}`),
		nil, fixtureTextNormalizer{},
	)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	digest := migrated.SHA256()
	projected, err := ProjectMigrationSource(
		source, target, mappings, migrated.CanonicalJSON(),
		digest[:], fixtureTextNormalizer{},
	)
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	if string(projected.CanonicalJSON()) !=
		`{"11111111-1111-4111-8111-111111111111":"ADA@EXAMPLE.COM"}` {
		t.Fatalf("unexpected source view: %s", projected.CanonicalJSON())
	}
}
