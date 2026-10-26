package recordkernel

import (
	"fmt"
	"github.com/google/uuid"
)

func compileRecordExpressions(
	schema StorageSchema,
	rules RecordRules,
	fields map[uuid.UUID]Field,
	limits BudgetLimits,
) ([]defaultValue, []calculation, []crossFieldRule, error) {
	defaults, defaultNodes, err := compileDefaults(rules.Defaults, fields, limits)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: defaults", err)
	}
	calculations, calculationNodes, err := compileCalculations(
		schema.Calculations, fields, limits,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: calculations", err)
	}
	crossRules, crossNodes, err := compileCrossFieldRules(
		rules.CrossFieldValidations, fields, limits,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: cross-field rules", err)
	}
	if defaultNodes+calculationNodes+crossNodes > limits.CalculationTotalASTNodes {
		return nil, nil, nil, fmt.Errorf("%w: total AST nodes", ErrCatalogInvalid)
	}
	if !engineFieldsSatisfiable(fields, defaults) {
		return nil, nil, nil, fmt.Errorf("%w: engine fields", ErrCatalogInvalid)
	}
	return defaults, calculations, crossRules, nil
}

func engineFieldsSatisfiable(
	fields map[uuid.UUID]Field,
	defaults []defaultValue,
) bool {
	hasDefault := make(map[uuid.UUID]struct{}, len(defaults))
	for _, value := range defaults {
		hasDefault[value.fieldID] = struct{}{}
	}
	for _, field := range fields {
		if field.WritePolicy == WriteEngine && field.Role != RoleCalculation &&
			!field.Nullable {
			if _, found := hasDefault[field.ID]; !found {
				return false
			}
		}
	}
	return true
}
