package profile

import (
	"strings"
	"testing"
)

func TestFieldValueTypeMatrix(t *testing.T) {
	minimum, maximum := int64(0), int64(10)
	integer := baseFieldInput(fieldOneID, "employee_number", LogicalInteger)
	integer.Validation = ValidationInput{IntegerMinimum: &minimum, IntegerMaximum: &maximum}
	decimal := baseFieldInput(fieldOneID, "credit_limit", LogicalDecimal)
	decimal.Validation = ValidationInput{DecimalPrecision: 4, DecimalScale: 2}
	enum := baseFieldInput(fieldOneID, "lifecycle", LogicalEnum)
	enum.Validation = ValidationInput{EnumValues: []string{"OPEN", "CLOSED"}}
	tests := []struct {
		name      string
		input     FieldInput
		valid     string
		canonical string
		invalid   string
	}{
		{"string", baseFieldInput(fieldOneID, "nickname", LogicalString), `""`, `""`, `"Cafe\u0301"`},
		{"integer", integer, `"10"`, `"10"`, `"11"`},
		{"decimal", decimal, `"12.3"`, `"12.30"`, `"12.345"`},
		{"boolean", baseFieldInput(fieldOneID, "newsletter", LogicalBoolean), `true`, `true`, `"true"`},
		{"date", baseFieldInput(fieldOneID, "start_date", LogicalDate), `"2026-07-16"`, `"2026-07-16"`, `"16-07-2026"`},
		{"timestamp", baseFieldInput(fieldOneID, "last_seen", LogicalTimestamp),
			`"2026-07-16T12:30:00.000000Z"`, `"2026-07-16T12:30:00.000000Z"`,
			`"2026-07-16T12:30:00Z"`},
		{"uuid", baseFieldInput(fieldOneID, "external_ref", LogicalUUID),
			`"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"`,
			`"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"`,
			`"AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA"`},
		{"enum", enum, `"OPEN"`, `"OPEN"`, `"open"`},
		{"json", baseFieldInput(fieldOneID, "preferences", LogicalJSON),
			`{"b":1,"a":"x"}`, `{"a":"x","b":1}`, `null`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			field := mustField(t, test.input)
			value, err := normalizeFieldValue(
				field, []byte(test.valid), fixtureTextNormalizer{},
			)
			if err != nil || string(value) != test.canonical {
				t.Fatalf("normalized = %s, %v", value, err)
			}
			if _, err := normalizeFieldValue(
				field, []byte(test.invalid), fixtureTextNormalizer{},
			); err == nil {
				t.Fatal("invalid typed value was accepted")
			}
		})
	}
}

func TestCollectionRejectsNullElementsAndPreservesEmptyList(t *testing.T) {
	input := baseFieldInput(fieldOneID, "aliases", LogicalString)
	input.Cardinality = CardinalityList
	input.Validation.MaximumItems = 2
	field := mustField(t, input)
	value, err := normalizeFieldValue(field, []byte(`[]`), fixtureTextNormalizer{})
	if err != nil || string(value) != `[]` {
		t.Fatalf("empty list = %s, %v", value, err)
	}
	for _, wire := range []string{`[null]`, `["a","b","c"]`, `["Cafe\u0301"]`} {
		if _, err := normalizeFieldValue(
			field, []byte(wire), fixtureTextNormalizer{},
		); err == nil {
			t.Fatalf("invalid collection %s was accepted", wire)
		}
	}
}

func TestJSONValueRejectsDuplicateAndNULMembers(t *testing.T) {
	field := mustField(t, baseFieldInput(fieldOneID, "preferences", LogicalJSON))
	for _, wire := range []string{
		`{"a":1,"a":2}`, `{"bad\u0000key":true}`, `{"value":"bad\u0000text"}`,
	} {
		if _, err := normalizeFieldValue(
			field, []byte(wire), fixtureTextNormalizer{},
		); err == nil {
			t.Fatalf("invalid JSON %s was accepted", wire)
		}
	}
}

func TestStringMaximumBytesCountsJSONUTF8WireBytes(t *testing.T) {
	input := baseFieldInput(fieldOneID, "display_name", LogicalString)
	input.Validation.MaximumBytes = 128
	field := mustField(t, input)

	withinLimit := `"` + strings.Repeat("é", 63) + `"`
	if _, err := normalizeFieldValue(
		field, []byte(withinLimit), fixtureTextNormalizer{},
	); err != nil {
		t.Fatalf("128-byte JSON string was rejected: %v", err)
	}

	overLimit := `"` + strings.Repeat("é", 64) + `"`
	if got := len([]byte(overLimit)); got != 130 {
		t.Fatalf("boundary fixture bytes = %d, want 130", got)
	}
	if _, err := normalizeFieldValue(
		field, []byte(overLimit), fixtureTextNormalizer{},
	); err == nil {
		t.Fatal("130-byte JSON string was accepted by a 128-byte field")
	}
}
