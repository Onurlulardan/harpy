package catalog

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func fieldCapabilities(
	fieldType api.FieldType,
	operators []string,
	sortable bool,
	cursorBytes int,
	searchable bool,
) (api.FieldCapabilities, error) {
	filters := make([]api.FilterCapabilityInput, len(operators))
	for index, value := range operators {
		operator, found := api.ParseOperator(value)
		if !found {
			return api.FieldCapabilities{}, ErrContract
		}
		filters[index] = api.FilterCapabilityInput{Operator: operator, CountSafe: true}
	}
	return api.NewFieldCapabilities(fieldType, api.FieldCapabilitiesInput{
		Selectable: true, Filters: filters, Sortable: sortable,
		CursorCanonicalBytes: cursorBytes, Searchable: searchable,
	})
}

func operatorNames(operators []api.Operator) ([]string, error) {
	result := make([]string, len(operators))
	for index, operator := range operators {
		if _, valid := operator.Arity(); !valid || operator.String() == "" {
			return nil, ErrContract
		}
		result[index] = operator.String()
	}
	return result, nil
}
