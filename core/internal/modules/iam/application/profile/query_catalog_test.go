package profile

import (
	"bytes"
	"testing"

	restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func TestQueryCatalogExposesOnlyClosedCustomCapabilities(t *testing.T) {
	text := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	listInput := baseFieldInput(fieldTwoID, "external_refs", LogicalUUID)
	listInput.Cardinality = CardinalityList
	list := mustField(t, listInput)
	jsonField := mustField(t, baseFieldInput(fieldThreeID, "preferences", LogicalJSON))
	retiredInput := baseFieldInput(fieldFourID, "old_label", LogicalString)
	retiredInput.State = FieldRetired
	retired := mustField(t, retiredInput)
	catalog, err := NewQueryCatalog(
		mustSchema(t, schemaOneID, 1, text, list, jsonField, retired),
	)
	if err != nil {
		t.Fatalf("new query catalog: %v", err)
	}

	assertQueryField(t, catalog, fieldOneID, restquery.LogicalTypeText,
		[]restquery.Operator{restquery.OperatorEq, restquery.OperatorIn,
			restquery.OperatorIsNull},
		[]restquery.Operator{restquery.OperatorGT, restquery.OperatorContains})
	assertQueryField(t, catalog, fieldTwoID, restquery.LogicalTypeUUIDList,
		[]restquery.Operator{restquery.OperatorContainsAny, restquery.OperatorContainsAll,
			restquery.OperatorIsNotNull},
		[]restquery.Operator{restquery.OperatorEq, restquery.OperatorIn})
	assertQueryField(t, catalog, fieldThreeID, restquery.LogicalTypeJSON,
		[]restquery.Operator{restquery.OperatorEq, restquery.OperatorNe,
			restquery.OperatorIsNull},
		[]restquery.Operator{restquery.OperatorIn, restquery.OperatorGT})
	if _, err := catalog.Field(fieldFourID); err == nil {
		t.Fatal("retired field appeared in the query catalog")
	}
}

func assertQueryField(
	t *testing.T,
	catalog QueryCatalog,
	id uuid.UUID,
	wantType restquery.LogicalType,
	allowed []restquery.Operator,
	rejected []restquery.Operator,
) {
	t.Helper()
	field, err := catalog.Field(id)
	if err != nil {
		t.Fatalf("resolve query field: %v", err)
	}
	kind, err := field.LogicalType()
	if err != nil || kind != wantType {
		t.Fatalf("query logical type = %v, %v", kind, err)
	}
	for _, operator := range allowed {
		if err := field.RequireOperator(operator); err != nil {
			t.Fatalf("allowed operator %s: %v", operator, err)
		}
	}
	for _, operator := range rejected {
		if err := field.RequireOperator(operator); err == nil {
			t.Fatalf("unsupported operator %s was accepted", operator)
		}
	}
	if field.RequireSort() == nil || field.RequireSearch() == nil {
		t.Fatal("custom sort or search capability was opened")
	}
}

func TestQueryCatalogCanonicalBytesAreOwned(t *testing.T) {
	field := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	catalog, err := NewQueryCatalog(mustSchema(t, schemaOneID, 1, field))
	if err != nil {
		t.Fatal(err)
	}
	first, _ := catalog.CanonicalJSON()
	second, _ := catalog.CanonicalJSON()
	first[0] = '!'
	if bytes.Equal(first, second) {
		t.Fatal("query catalog canonical bytes escaped ownership")
	}
}

func TestRetainedQueryCatalogAcceptsOnlyPinnedHistoricalSchema(t *testing.T) {
	field := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	retired, err := NewPublishedSchema(SchemaInput{
		ID: schemaOneID, Resource: ResourceUser, Version: 1,
		State: SchemaRetired, Fields: []Field{field},
	})
	if err != nil {
		t.Fatalf("retired schema: %v", err)
	}
	if _, err := NewQueryCatalog(retired); err == nil {
		t.Fatal("active query catalog accepted a retired schema")
	}
	retained, err := NewRetainedQueryCatalog(retired)
	if err != nil {
		t.Fatalf("retained query catalog: %v", err)
	}
	if retained.Validate() != nil {
		t.Fatal("retained query catalog is invalid")
	}
	active, err := NewQueryCatalog(mustSchema(t, schemaOneID, 1, field))
	if err != nil {
		t.Fatal(err)
	}
	retainedHash, retainedErr := retained.Hash()
	activeHash, activeErr := active.Hash()
	if retainedErr != nil || activeErr != nil || retainedHash != activeHash {
		t.Fatal("retained query catalog drifted from its published digest")
	}
}

func TestQueryFieldCarriesImmutableTypeAndClassification(t *testing.T) {
	minimum, maximum := int64(-12), int64(42)
	input := baseFieldInput(fieldOneID, "score", LogicalInteger)
	input.Classification = ClassificationConfidential
	input.Validation.MaximumBytes = 128
	input.Validation.IntegerMinimum = &minimum
	input.Validation.IntegerMaximum = &maximum
	catalog, err := NewQueryCatalog(mustSchema(t, schemaOneID, 1, mustField(t, input)))
	if err != nil {
		t.Fatal(err)
	}
	field, err := catalog.Field(fieldOneID)
	if err != nil {
		t.Fatal(err)
	}
	classification, err := field.Classification()
	if err != nil || classification != ClassificationConfidential {
		t.Fatalf("classification = %q, %v", classification, err)
	}
	maximumBytes, maximumBytesErr := field.MaximumBytes()
	if maximumBytesErr != nil || maximumBytes != 128 {
		t.Fatalf("maximum bytes = %d, %v", maximumBytes, maximumBytesErr)
	}
	fieldType, err := field.FieldType()
	profile, present, profileErr := fieldType.Integer()
	actualMinimum, minimumErr := profile.Minimum()
	actualMaximum, maximumErr := profile.Maximum()
	if err != nil || profileErr != nil || !present || minimumErr != nil ||
		maximumErr != nil || actualMinimum != minimum || actualMaximum != maximum {
		t.Fatal("query field lost its immutable integer contract")
	}
}
