package recordkernel

import (
	"github.com/google/uuid"
)

func compileDefaults(
	rules []DefaultRule,
	fields map[uuid.UUID]Field,
	limits BudgetLimits,
) ([]defaultValue, uint64, error) {
	values := make(map[uuid.UUID]defaultValue, len(rules))
	dependencies := make(map[uuid.UUID]map[uuid.UUID]struct{}, len(rules))
	var total uint64
	for _, rule := range rules {
		fieldID, err := parseStableID(rule.FieldID)
		field, found := fields[fieldID]
		if err != nil || !found || field.Role != RoleStored {
			return nil, 0, ErrCatalogInvalid
		}
		if _, duplicate := values[fieldID]; duplicate {
			return nil, 0, ErrCatalogInvalid
		}
		compiled, nodes, refs, err := compileTypedExpression(
			rule.Expression, field.TypeID, fields, 1, limits,
		)
		if err != nil {
			return nil, 0, err
		}
		for ref := range refs {
			if fields[ref].Role != RoleStored {
				return nil, 0, ErrCatalogInvalid
			}
		}
		values[fieldID] = defaultValue{fieldID, compiled}
		dependencies[fieldID] = refs
		total += nodes
	}
	order, err := topologicalOrder(defaultKeys(values), dependencies)
	if err != nil {
		return nil, 0, err
	}
	result := make([]defaultValue, len(order))
	for index, fieldID := range order {
		result[index] = values[fieldID]
	}
	return result, total, nil
}

func defaultKeys(values map[uuid.UUID]defaultValue) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(values))
	for fieldID := range values {
		result = append(result, fieldID)
	}
	return result
}
