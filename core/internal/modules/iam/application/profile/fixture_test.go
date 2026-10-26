package profile

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

var (
	fieldOneID   = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	fieldTwoID   = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	fieldThreeID = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	fieldFourID  = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	schemaOneID  = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	schemaTwoID  = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
)

type fixtureTextNormalizer struct{}

func (fixtureTextNormalizer) NormalizeEmailV1(value string) (string, error) {
	return strings.ToLower(strings.TrimSpace(value)), nil
}

func (fixtureTextNormalizer) NormalizePortalV1(value string) (string, error) {
	return strings.ToLower(strings.TrimSpace(value)), nil
}

func baseFieldInput(id uuid.UUID, key string, kind LogicalType) FieldInput {
	return FieldInput{
		ID: id, Resource: ResourceUser, Key: key, Type: kind,
		Cardinality: CardinalitySingle, Classification: ClassificationPublic,
		Normalization: NormalizationNoneV1, State: FieldActive,
	}
}

func mustField(t *testing.T, input FieldInput) Field {
	t.Helper()
	field, err := NewField(input)
	if err != nil {
		t.Fatalf("new field: %v", err)
	}
	return field
}

func mustSchema(t *testing.T, id uuid.UUID, version int64, fields ...Field) Schema {
	t.Helper()
	schema, err := NewPublishedSchema(SchemaInput{
		ID: id, Resource: ResourceUser, Version: version,
		State: SchemaActive, Fields: fields,
	})
	if err != nil {
		t.Fatalf("new schema: %v", err)
	}
	return schema
}

func mustDocument(t *testing.T, schema Schema, wire string) Document {
	t.Helper()
	document, err := DecodeDocument(schema, []byte(wire), nil, fixtureTextNormalizer{})
	if err != nil {
		t.Fatalf("decode document: %v", err)
	}
	return document
}
