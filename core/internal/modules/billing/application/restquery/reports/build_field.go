package reports

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func buildField(spec fieldSpec) (api.FieldDefinition, error) {
	kind, typeErr := fieldType(spec)
	storage, storageErr := fieldStorage(spec)
	filters, filterErr := filterCapabilities(spec.Filters)
	aggregates, aggregateErr := aggregateCapabilities(spec.Aggregates)
	if typeErr != nil || storageErr != nil || filterErr != nil || aggregateErr != nil {
		return api.FieldDefinition{}, ErrContract
	}
	capabilities, err := api.NewFieldCapabilities(kind, api.FieldCapabilitiesInput{
		Selectable: true, Filters: filters, Sortable: spec.Sortable,
		CursorCanonicalBytes: spec.Cursor, Groupable: spec.Groupable,
		Aggregates: aggregates,
	})
	if err != nil {
		return api.FieldDefinition{}, ErrContract
	}
	field, err := api.NewFieldDefinition(
		spec.ID, spec.Name, kind, storage, capabilities,
	)
	if err != nil {
		return api.FieldDefinition{}, ErrContract
	}
	return field, nil
}

func filterCapabilities(names []string) ([]api.FilterCapabilityInput, error) {
	result := make([]api.FilterCapabilityInput, len(names))
	for index, name := range names {
		operator, found := api.ParseOperator(name)
		if !found {
			return nil, ErrContract
		}
		result[index] = api.FilterCapabilityInput{Operator: operator, CountSafe: true}
	}
	return result, nil
}

func aggregateCapabilities(names []string) ([]api.AggregateFunction, error) {
	result := make([]api.AggregateFunction, len(names))
	for index, name := range names {
		function, found := api.ParseAggregateFunction(name)
		if !found {
			return nil, ErrContract
		}
		result[index] = function
	}
	return result, nil
}
