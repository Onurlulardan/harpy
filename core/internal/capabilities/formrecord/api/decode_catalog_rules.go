package recordkernel

import (
	"github.com/google/uuid"
)

func decodeCatalogRules(
	document catalogAttestationDocument,
	fields map[uuid.UUID]Field,
	limits BudgetLimits,
) ([]defaultValue, []calculation, []crossFieldRule, error) {
	defaults, err := decodeAssignments(
		document.Defaults, RoleStored, fields, limits,
	)
	if err != nil {
		return nil, nil, nil, err
	}
	calculations, err := decodeAssignments(
		document.Calculations, RoleCalculation, fields, limits,
	)
	if err != nil || !allCalculationsPresent(fields, calculations.calculations) {
		return nil, nil, nil, ErrCatalogInvalid
	}
	crossRules, crossNodes, err := decodeCrossRules(
		document.CrossFieldRules, fields, limits,
	)
	if err != nil || defaults.nodes+calculations.nodes+crossNodes >
		limits.CalculationTotalASTNodes ||
		!engineFieldsSatisfiable(fields, defaults.defaults) {
		return nil, nil, nil, ErrCatalogInvalid
	}
	return defaults.defaults, calculations.calculations, crossRules, nil
}

func allCalculationsPresent(
	fields map[uuid.UUID]Field,
	values []calculation,
) bool {
	present := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		present[value.fieldID] = struct{}{}
	}
	for fieldID, field := range fields {
		if field.Role == RoleCalculation {
			if _, found := present[fieldID]; !found {
				return false
			}
		}
	}
	return true
}
