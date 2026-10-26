package schemaexecutor

import (
	"testing"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func TestValidateBackfillSameFieldAcrossGenerations(t *testing.T) {
	fieldID := "00000000-0000-4000-8000-000000000006"
	mapping := backfillTestMapping(t, fieldID, 7, 8)
	arguments := schemav1.BackfillArguments{
		SourceFieldID: fieldID, SourceGeneration: 7,
		TargetFieldID: fieldID, TargetGeneration: 8,
		Converter: schemav1.ConvertIdentity,
	}
	if err := validateBackfillArguments(mapping, arguments); err != nil {
		t.Fatalf("validate generation backfill: %v", err)
	}
}

func TestValidateBackfillRejectsSameGeneration(t *testing.T) {
	fieldID := "00000000-0000-4000-8000-000000000006"
	mapping := backfillTestMapping(t, fieldID, 7)
	arguments := schemav1.BackfillArguments{
		SourceFieldID: fieldID, SourceGeneration: 7,
		TargetFieldID: fieldID, TargetGeneration: 7,
		Converter: schemav1.ConvertIdentity,
	}
	if err := validateBackfillArguments(mapping, arguments); err == nil {
		t.Fatal("expected same-generation backfill rejection")
	}
}

func TestValidateBackfillRequiresExactGeneration(t *testing.T) {
	fieldID := "00000000-0000-4000-8000-000000000006"
	mapping := backfillTestMapping(t, fieldID, 7)
	arguments := schemav1.BackfillArguments{
		SourceFieldID: fieldID, SourceGeneration: 7,
		TargetFieldID: fieldID, TargetGeneration: 8,
		Converter: schemav1.ConvertIdentity,
	}
	if err := validateBackfillArguments(mapping, arguments); err == nil {
		t.Fatal("expected missing target-generation rejection")
	}
}

func TestValidateBackfillAllowsDifferentFieldsAtSameGeneration(t *testing.T) {
	sourceID := "00000000-0000-4000-8000-000000000006"
	targetID := "00000000-0000-4000-8000-000000000007"
	sourceName, err := deriveFieldIdentifier(sourceID, 7)
	if err != nil {
		t.Fatal(err)
	}
	targetName, err := deriveFieldIdentifier(targetID, 7)
	if err != nil {
		t.Fatal(err)
	}
	mapping := schemav1.IdentifierMapping{Fields: []schemav1.FieldIdentifier{
		{LogicalID: sourceID, Generation: 7, PhysicalName: sourceName},
		{LogicalID: targetID, Generation: 7, PhysicalName: targetName},
	}}
	arguments := schemav1.BackfillArguments{
		SourceFieldID: sourceID, SourceGeneration: 7,
		TargetFieldID: targetID, TargetGeneration: 7,
		Converter: schemav1.ConvertIdentity,
	}
	if err := validateBackfillArguments(mapping, arguments); err != nil {
		t.Fatalf("validate same-generation cross-field backfill: %v", err)
	}
}

func backfillTestMapping(
	t *testing.T,
	fieldID string,
	generations ...int64,
) schemav1.IdentifierMapping {
	t.Helper()
	fields := make([]schemav1.FieldIdentifier, len(generations))
	for index, generation := range generations {
		name, err := deriveFieldIdentifier(fieldID, generation)
		if err != nil {
			t.Fatal(err)
		}
		fields[index] = schemav1.FieldIdentifier{
			LogicalID: fieldID, Generation: generation, PhysicalName: name,
		}
	}
	return schemav1.IdentifierMapping{Fields: fields}
}
