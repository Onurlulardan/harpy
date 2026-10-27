package profile

import (
	"errors"
	"testing"
)

func TestSelectMigrationMappingsFiltersFieldsAbsentFromHistoricalSource(t *testing.T) {
	first := mustField(t, baseFieldInput(fieldOneID, "display_name", LogicalString))
	second := mustField(t, baseFieldInput(fieldTwoID, "nickname", LogicalString))
	historical := mustSchema(t, schemaOneID, 1, first)
	target := mustSchema(t, schemaTwoID, 3, first, second)
	mappings := []MigrationMapping{
		{SourceFieldID: fieldOneID, TargetFieldID: fieldOneID, Converter: MigrationCopy},
		{SourceFieldID: fieldTwoID, TargetFieldID: fieldTwoID, Converter: MigrationCopy},
	}
	selected, err := SelectMigrationMappingsForSource(historical, target, mappings)
	if err != nil || len(selected) != 1 || selected[0].SourceFieldID != fieldOneID {
		t.Fatalf("selected mappings = %+v, %v", selected, err)
	}
}

func TestSelectMigrationMappingsRequiresHistoricalCoverageForRequiredTarget(t *testing.T) {
	first := mustField(t, baseFieldInput(fieldOneID, "display_name", LogicalString))
	requiredInput := baseFieldInput(fieldTwoID, "nickname", LogicalString)
	requiredInput.Required = true
	required := mustField(t, requiredInput)
	historical := mustSchema(t, schemaOneID, 1, first)
	target := mustSchema(t, schemaTwoID, 3, first, required)
	mappings := []MigrationMapping{
		{SourceFieldID: fieldOneID, TargetFieldID: fieldOneID, Converter: MigrationCopy},
		{SourceFieldID: fieldTwoID, TargetFieldID: fieldTwoID, Converter: MigrationCopy},
	}
	if _, err := SelectMigrationMappingsForSource(historical, target, mappings); !errors.Is(
		err, ErrSchemaInvalid,
	) {
		t.Fatalf("required target coverage error = %v", err)
	}
}

func TestSelectMigrationMappingsRejectsInvalidFullContract(t *testing.T) {
	field := mustField(t, baseFieldInput(fieldOneID, "display_name", LogicalString))
	source := mustSchema(t, schemaOneID, 1, field)
	target := mustSchema(t, schemaTwoID, 2, field)
	tests := [][]MigrationMapping{
		{{SourceFieldID: fieldOneID, TargetFieldID: fieldOneID, Converter: "unknown"}},
		{
			{SourceFieldID: fieldOneID, TargetFieldID: fieldOneID, Converter: MigrationCopy},
			{SourceFieldID: fieldOneID, TargetFieldID: fieldTwoID, Converter: MigrationCopy},
		},
	}
	for _, mappings := range tests {
		if _, err := SelectMigrationMappingsForSource(source, target, mappings); !errors.Is(
			err, ErrSchemaInvalid,
		) {
			t.Fatalf("invalid mapping contract error = %v", err)
		}
	}
}

func TestSelectMigrationMappingsAllowsEmptyHistoricalSubset(t *testing.T) {
	historical, err := NewPublishedSchema(SchemaInput{
		ID: schemaOneID, Resource: ResourceUser, Version: 1,
		State: SchemaRetired, Fields: []Field{},
	})
	if err != nil {
		t.Fatal(err)
	}
	optional := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	target := mustSchema(t, schemaTwoID, 3, optional)
	selected, err := SelectMigrationMappingsForSource(
		historical, target, []MigrationMapping{{
			SourceFieldID: optional.ID(), TargetFieldID: optional.ID(),
			Converter: MigrationCopy,
		}},
	)
	if err != nil || len(selected) != 0 {
		t.Fatalf("empty historical subset = %+v, %v", selected, err)
	}
}
