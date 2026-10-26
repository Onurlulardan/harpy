package presentation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func (state *validationState) validateVisibility(path string, value *Visibility) error {
	if value == nil {
		return nil
	}
	if value.Match != "all" && value.Match != "any" ||
		len(value.Conditions) < 1 || len(value.Conditions) > maximumConditions {
		return invalid(path, "match or condition budget")
	}
	for index, condition := range value.Conditions {
		conditionPath := fmt.Sprintf("%s.conditions[%d]", path, index)
		if _, exists := state.fields[condition.FieldID]; !exists {
			return invalid(conditionPath+".field_id", "unknown")
		}
		if err := validateCondition(condition, conditionPath); err != nil {
			return err
		}
	}
	return nil
}

func validateCondition(condition Condition, path string) error {
	if condition.Operator == "is_null" || condition.Operator == "is_not_null" {
		if condition.Value != nil {
			return invalid(path+".value", "null operator takes no operand")
		}
		return nil
	}
	if !valueOperator(condition.Operator) || len(condition.Value) == 0 ||
		len(condition.Value) > maximumOperandBytes {
		return invalid(path, "operator or operand")
	}
	var operand any
	decoder := json.NewDecoder(bytes.NewReader(condition.Value))
	decoder.UseNumber()
	if decoder.Decode(&operand) != nil {
		return invalid(path+".value", "operand encoding")
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return invalid(path+".value", "trailing operand")
	}
	if condition.Operator == "in" || condition.Operator == "not_in" {
		values, ok := operand.([]any)
		if !ok || len(values) < 1 || len(values) > maximumOperandItems {
			return invalid(path+".value", "list operand")
		}
		for _, value := range values {
			if !scalarOperand(value) {
				return invalid(path+".value", "scalar list operand")
			}
		}
		return nil
	}
	if !scalarOperand(operand) {
		return invalid(path+".value", "scalar operand")
	}
	return nil
}

func scalarOperand(value any) bool {
	switch value.(type) {
	case string, bool, json.Number:
		return true
	default:
		return false
	}
}

func valueOperator(value string) bool {
	switch value {
	case "eq", "ne", "gt", "gte", "lt", "lte", "contains", "in", "not_in":
		return true
	default:
		return false
	}
}
