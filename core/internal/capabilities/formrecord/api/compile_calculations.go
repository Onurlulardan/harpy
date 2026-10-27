package recordkernel

import (
	"github.com/google/uuid"
)

func compileCalculations(
	rules []CalculationDefinition,
	fields map[uuid.UUID]Field,
	limits BudgetLimits,
) ([]calculation, uint64, error) {
	if uint64(len(rules)) > limits.CalculationFields {
		return nil, 0, ErrCatalogInvalid
	}
	values := make(map[uuid.UUID]calculation, len(rules))
	dependencies := make(map[uuid.UUID]map[uuid.UUID]struct{}, len(rules))
	var total uint64
	for _, rule := range rules {
		fieldID, err := parseStableID(rule.FieldID)
		field, found := fields[fieldID]
		if err != nil || !found || field.Role != RoleCalculation {
			return nil, 0, ErrCatalogInvalid
		}
		if _, duplicate := values[fieldID]; duplicate {
			return nil, 0, ErrCatalogInvalid
		}
		compiled, nodes, refs, err := compileLegacyExpression(
			rule.Expression, field.TypeID, fields, 1, limits,
		)
		if err != nil {
			return nil, 0, err
		}
		values[fieldID] = calculation{fieldID, compiled}
		dependencies[fieldID] = refs
		total += nodes
	}
	for fieldID, field := range fields {
		if field.Role == RoleCalculation {
			if _, found := values[fieldID]; !found {
				return nil, 0, ErrCatalogInvalid
			}
		}
	}
	order, err := topologicalOrder(calculationKeys(values), dependencies)
	if err != nil {
		return nil, 0, err
	}
	result := make([]calculation, len(order))
	for index, fieldID := range order {
		result[index] = values[fieldID]
	}
	return result, total, nil
}

func calculationKeys(values map[uuid.UUID]calculation) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(values))
	for fieldID := range values {
		result = append(result, fieldID)
	}
	return result
}
