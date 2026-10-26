package v1query

import (
	"fmt"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func validateFilterFieldBindings(
	inputs map[string]FilterFieldBinding,
) (map[string]fieldBinding, error) {
	result := make(map[string]fieldBinding, len(inputs))
	for raw, input := range inputs {
		shapeValid := input.Shape == FilterShapeScalar ||
			input.Shape == FilterShapeCollection
		if strings.ContainsAny(raw, "[]") || !shapeValid ||
			!validFieldBinding(raw, input.Name, input.AliasID) {
			return nil, fmt.Errorf("verified V1 filter-field binding is invalid")
		}
		result[raw] = fieldBinding{
			name: input.Name, shape: input.Shape, aliasID: input.AliasID,
		}
	}
	return result, nil
}

func validateSortFieldBindings(
	inputs map[string]SortFieldBinding,
) (map[string]fieldBinding, error) {
	result := make(map[string]fieldBinding, len(inputs))
	for raw, input := range inputs {
		if !validFieldBinding(raw, input.Name, input.AliasID) {
			return nil, fmt.Errorf("verified V1 sort-field binding is invalid")
		}
		result[raw] = fieldBinding{
			name: input.Name, shape: FilterShapeScalar, aliasID: input.AliasID,
		}
	}
	return result, nil
}

func validFieldBinding(raw, canonical, aliasID string) bool {
	_, fieldErr := api.NewRequestSort(canonical, api.DirectionAscending)
	return validWireName(raw) && fieldErr == nil &&
		validOptionalAlias(raw != canonical, aliasID)
}
