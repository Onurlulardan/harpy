package recordkernel

import (
	"github.com/google/uuid"
)

type decodedAssignments struct {
	defaults     []defaultValue
	calculations []calculation
	nodes        uint64
}

func decodeAssignments(
	documents []catalogAssignment,
	role Role,
	fields map[uuid.UUID]Field,
	limits BudgetLimits,
) (decodedAssignments, error) {
	if documents == nil || len(documents) > len(fields) ||
		role == RoleCalculation && uint64(len(documents)) >
			limits.CalculationFields {
		return decodedAssignments{}, ErrCatalogInvalid
	}
	compiled := make(map[uuid.UUID]expression, len(documents))
	dependencies := make(map[uuid.UUID]map[uuid.UUID]struct{}, len(documents))
	order := make([]uuid.UUID, len(documents))
	var total uint64
	for index, document := range documents {
		fieldID, err := parseCatalogID(document.FieldID)
		field, found := fields[fieldID]
		if err != nil || !found || field.Role != role {
			return decodedAssignments{}, ErrCatalogInvalid
		}
		if _, duplicate := compiled[fieldID]; duplicate {
			return decodedAssignments{}, ErrCatalogInvalid
		}
		expression, nodes, refs, err := decodeCatalogExpression(
			document.Expression, field.TypeID, fields, limits,
		)
		if err != nil {
			return decodedAssignments{}, err
		}
		if role == RoleStored && !storedReferences(refs, fields) {
			return decodedAssignments{}, ErrCatalogInvalid
		}
		compiled[fieldID], dependencies[fieldID], order[index] = expression, refs, fieldID
		total += nodes
	}
	expected, err := topologicalOrder(append([]uuid.UUID(nil), order...), dependencies)
	if err != nil || !sameUUIDOrder(order, expected) {
		return decodedAssignments{}, ErrCatalogInvalid
	}
	result := decodedAssignments{nodes: total}
	if role == RoleStored {
		result.defaults = make([]defaultValue, len(order))
		for index, fieldID := range order {
			result.defaults[index] = defaultValue{fieldID, compiled[fieldID]}
		}
	} else {
		result.calculations = make([]calculation, len(order))
		for index, fieldID := range order {
			result.calculations[index] = calculation{fieldID, compiled[fieldID]}
		}
	}
	return result, nil
}

func storedReferences(
	refs map[uuid.UUID]struct{},
	fields map[uuid.UUID]Field,
) bool {
	for fieldID := range refs {
		if fields[fieldID].Role != RoleStored {
			return false
		}
	}
	return true
}

func sameUUIDOrder(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
