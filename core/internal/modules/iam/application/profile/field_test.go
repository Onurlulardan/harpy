package profile

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
)

func TestNewFieldRejectsOpenOrSecuritySemanticShapes(t *testing.T) {
	decimal := baseFieldInput(fieldOneID, "credit_limit", LogicalDecimal)
	decimal.Validation = ValidationInput{DecimalPrecision: 12, DecimalScale: 2}
	valid := mustField(t, decimal)
	if valid.Type() != LogicalDecimal || valid.Cardinality() != CardinalitySingle {
		t.Fatal("valid field semantic was not preserved")
	}

	tests := map[string]FieldInput{
		"nil id":              baseFieldInput(uuid.Nil, "nickname", LogicalString),
		"unknown resource":    baseFieldInput(fieldOneID, "nickname", LogicalString),
		"invalid key":         baseFieldInput(fieldOneID, "DisplayName", LogicalString),
		"reserved core":       baseFieldInput(fieldOneID, "email", LogicalString),
		"security term":       baseFieldInput(fieldOneID, "password_hint", LogicalString),
		"unsupported list":    baseFieldInput(fieldOneID, "scores", LogicalInteger),
		"decimal unspecified": baseFieldInput(fieldOneID, "credit_limit", LogicalDecimal),
	}
	tests["unknown resource"] = func() FieldInput {
		input := tests["unknown resource"]
		input.Resource = "unknown"
		return input
	}()
	tests["unsupported list"] = func() FieldInput {
		input := tests["unsupported list"]
		input.Cardinality = CardinalityList
		return input
	}()
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := NewField(input); err == nil {
				t.Fatal("invalid field was accepted")
			}
		})
	}
}

func TestFieldOwnsAndCanonicalizesValidationContract(t *testing.T) {
	values := []string{"OPEN", "CLOSED"}
	input := baseFieldInput(fieldOneID, "lifecycle", LogicalEnum)
	input.Validation = ValidationInput{MaximumBytes: 64, EnumValues: values}
	field := mustField(t, input)
	before := field.ValidationJSON()
	values[0] = "MUTATED"
	returned := field.ValidationJSON()
	returned[0] = '!'
	if !bytes.Equal(before, field.ValidationJSON()) ||
		!bytes.Contains(before, []byte(`"enum_values":["CLOSED","OPEN"]`)) {
		t.Fatal("validation contract is mutable or non-canonical")
	}
}

func TestNewFieldAcceptsOnlyRESTQueryRepresentableLists(t *testing.T) {
	for _, kind := range []LogicalType{LogicalString, LogicalUUID} {
		input := baseFieldInput(fieldOneID, "external_refs", kind)
		input.Cardinality = CardinalityList
		input.Validation.MaximumItems = 10
		if _, err := NewField(input); err != nil {
			t.Fatalf("list kind %s: %v", kind, err)
		}
	}
}
