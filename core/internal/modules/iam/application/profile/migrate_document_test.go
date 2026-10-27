package profile

import (
	"testing"

	"github.com/google/uuid"
)

func TestMigrateDocumentCopiesAndNormalizesThroughTargetContract(t *testing.T) {
	sourceField := mustField(t, baseFieldInput(fieldOneID, "display_name", LogicalString))
	source := mustSchema(t, schemaOneID, 1, sourceField)
	replacement := baseFieldInput(
		uuid.MustParse("55555555-5555-4555-8555-555555555555"),
		"contact_alias_v2", LogicalString,
	)
	replacement.Normalization = NormalizationEmailV1
	targetField, err := NewField(replacement)
	if err != nil {
		t.Fatalf("target field: %v", err)
	}
	retiredInput := baseFieldInput(fieldOneID, "display_name", LogicalString)
	retiredInput.State = FieldRetired
	retiredSource := mustField(t, retiredInput)
	target, err := NewPublishedSchema(SchemaInput{
		ID:       uuid.MustParse("66666666-6666-4666-8666-666666666666"),
		Resource: source.Resource(), Version: source.Version() + 1,
		State: SchemaActive, Fields: []Field{retiredSource, targetField},
	})
	if err != nil {
		t.Fatalf("target schema: %v", err)
	}
	value := `{"11111111-1111-4111-8111-111111111111":"ADA@EXAMPLE.COM"}`
	result, err := MigrateDocument(source, target, []MigrationMapping{{
		SourceFieldID: source.Fields()[0].ID(), TargetFieldID: targetField.ID(),
		Converter: MigrationNormalize,
	}}, []byte(value), nil, fixtureTextNormalizer{})
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if result.SchemaID() != target.ID() ||
		string(result.CanonicalJSON()) != `{"11111111-1111-4111-8111-111111111111":"ADA@EXAMPLE.COM","55555555-5555-4555-8555-555555555555":"ada@example.com"}` {
		t.Fatalf("unexpected target document: %s", result.CanonicalJSON())
	}
	retiredTarget, err := RetiredField(targetField)
	if err != nil {
		t.Fatal(err)
	}
	nextInput := baseFieldInput(
		uuid.MustParse("77777777-7777-4777-8777-777777777777"),
		"contact_alias_v3", LogicalString,
	)
	nextInput.Normalization = NormalizationEmailV1
	nextField := mustField(t, nextInput)
	next, err := NewPublishedSchema(SchemaInput{
		ID:       uuid.MustParse("88888888-8888-4888-8888-888888888888"),
		Resource: source.Resource(), Version: target.Version() + 1,
		State:  SchemaActive,
		Fields: []Field{retiredSource, retiredTarget, nextField},
	})
	if err != nil {
		t.Fatal(err)
	}
	resultSHA := result.SHA256()
	nextResult, err := MigrateDocument(target, next, []MigrationMapping{{
		SourceFieldID: targetField.ID(), TargetFieldID: nextField.ID(),
		Converter: MigrationNormalize,
	}}, result.CanonicalJSON(), resultSHA[:], fixtureTextNormalizer{})
	if err != nil || string(nextResult.CanonicalJSON()) !=
		`{"11111111-1111-4111-8111-111111111111":"ADA@EXAMPLE.COM","55555555-5555-4555-8555-555555555555":"ada@example.com","77777777-7777-4777-8777-777777777777":"ada@example.com"}` {
		t.Fatalf("retired compatibility value was lost: %s/%v",
			nextResult.CanonicalJSON(), err)
	}
}

func TestMigrateDocumentRejectsDuplicateAndInvalidConverterMappings(t *testing.T) {
	field := mustField(t, baseFieldInput(fieldOneID, "display_name", LogicalString))
	source := mustSchema(t, schemaOneID, 1, field)
	target := validSchemaWithID(t, uuid.MustParse("77777777-7777-4777-8777-777777777777"))
	fieldID := source.Fields()[0].ID()
	cases := [][]MigrationMapping{
		{{SourceFieldID: fieldID, TargetFieldID: fieldID, Converter: "unknown"}},
		{
			{SourceFieldID: fieldID, TargetFieldID: fieldID, Converter: MigrationCopy},
			{SourceFieldID: fieldID, TargetFieldID: fieldID, Converter: MigrationCopy},
		},
	}
	for _, mappings := range cases {
		if _, err := MigrateDocument(
			source, target, mappings,
			[]byte(`{"11111111-1111-4111-8111-111111111111":"Ada"}`),
			nil, fixtureTextNormalizer{},
		); err == nil {
			t.Fatal("invalid mappings were accepted")
		}
	}
}

func TestMigrateDocumentAllowsEmptyHistoricalMappingWithoutDataLoss(t *testing.T) {
	source, err := NewPublishedSchema(SchemaInput{
		ID: schemaOneID, Resource: ResourceUser, Version: 1,
		State: SchemaRetired, Fields: []Field{},
	})
	if err != nil {
		t.Fatal(err)
	}
	optional := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	target := mustSchema(t, schemaTwoID, 3, optional)
	result, err := MigrateDocument(
		source, target, nil, []byte(`{}`), nil, fixtureTextNormalizer{},
	)
	if err != nil || string(result.CanonicalJSON()) != `{}` {
		t.Fatalf("empty historical migration = %s, %v", result.CanonicalJSON(), err)
	}
}

func TestMigrateDocumentRejectsEmptyMappingThatCouldDropActiveData(t *testing.T) {
	field := mustField(t, baseFieldInput(fieldOneID, "display_name", LogicalString))
	source := mustSchema(t, schemaOneID, 1, field)
	target := mustSchema(t, schemaTwoID, 3, field)
	if _, err := MigrateDocument(
		source, target, nil,
		[]byte(`{"11111111-1111-4111-8111-111111111111":"Ada"}`), nil,
		fixtureTextNormalizer{},
	); err == nil {
		t.Fatal("empty mapping dropped an active source field")
	}
}

func validSchemaWithID(t *testing.T, id uuid.UUID) Schema {
	t.Helper()
	field := mustField(t, baseFieldInput(fieldOneID, "display_name", LogicalString))
	source := mustSchema(t, schemaOneID, 1, field)
	result, err := NewPublishedSchema(SchemaInput{
		ID: id, Resource: source.Resource(), Version: source.Version() + 1,
		State: SchemaActive, Fields: source.Fields(),
	})
	if err != nil {
		t.Fatalf("schema: %v", err)
	}
	return result
}
