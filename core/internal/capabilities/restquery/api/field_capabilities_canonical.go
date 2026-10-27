package api

import (
	"cmp"
	"slices"
)

func canonicalFilterCapabilities(
	inputs []FilterCapabilityInput,
) ([]FilterCapability, error) {
	result := make([]FilterCapability, 0, len(inputs))
	for _, input := range inputs {
		capability, err := newFilterCapability(input)
		if err != nil {
			return nil, errInvalidFieldCapabilities
		}
		result = append(result, capability)
	}
	slices.SortFunc(result, func(left, right FilterCapability) int {
		return cmp.Compare(left.operator, right.operator)
	})
	for index := 1; index < len(result); index++ {
		if result[index-1].operator == result[index].operator {
			return nil, errInvalidFieldCapabilities
		}
	}
	return result, nil
}

func canonicalAggregateCapabilities(
	inputs []AggregateFunction,
) ([]AggregateFunction, error) {
	result := append(make([]AggregateFunction, 0, len(inputs)), inputs...)
	slices.Sort(result)
	for index, function := range result {
		if !function.valid() || function == AggregateFunctionCount ||
			index > 0 && result[index-1] == function {
			return nil, errInvalidFieldCapabilities
		}
	}
	return result, nil
}

func validFilterCapabilities(capabilities FieldCapabilities) bool {
	previous := Operator(0)
	for _, filter := range capabilities.filters {
		if filter.Validate() != nil || filter.operator <= previous ||
			!filterOperatorAllowed(capabilities, filter.operator) {
			return false
		}
		previous = filter.operator
	}
	return true
}

func validAggregateCapabilities(capabilities FieldCapabilities) bool {
	previous := AggregateFunction(0)
	for _, function := range capabilities.aggregates {
		if function <= previous || !aggregateFunctionAllowed(capabilities, function) {
			return false
		}
		previous = function
	}
	return true
}
