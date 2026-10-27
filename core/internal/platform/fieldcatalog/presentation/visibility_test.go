package presentation

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestVisibilityOperatorOperandContract(t *testing.T) {
	for _, condition := range []Condition{
		{Operator: "eq", Value: json.RawMessage(`"x"`)},
		{Operator: "ne", Value: json.RawMessage(`true`)},
		{Operator: "gt", Value: json.RawMessage(`1`)},
		{Operator: "gte", Value: json.RawMessage(`1`)},
		{Operator: "lt", Value: json.RawMessage(`1`)},
		{Operator: "lte", Value: json.RawMessage(`1`)},
		{Operator: "contains", Value: json.RawMessage(`"x"`)},
		{Operator: "in", Value: json.RawMessage(`[1,"x"]`)},
		{Operator: "not_in", Value: json.RawMessage(`[false,"x"]`)},
		{Operator: "is_null"},
		{Operator: "is_not_null"},
	} {
		if err := validateCondition(condition, "$.condition"); err != nil {
			t.Fatalf("operator %s: %v", condition.Operator, err)
		}
	}
}

func TestVisibilityRejectsStructuredOrTrailingOperand(t *testing.T) {
	for _, value := range []json.RawMessage{
		json.RawMessage(`{"sql":"x"}`),
		json.RawMessage(`[[1]]`),
		json.RawMessage(`"x" true`),
		json.RawMessage(`null`),
	} {
		condition := Condition{Operator: "in", Value: value}
		if err := validateCondition(condition, "$.condition"); !errors.Is(err, ErrInvalid) {
			t.Fatalf("operand %s: %v", value, err)
		}
	}
}
