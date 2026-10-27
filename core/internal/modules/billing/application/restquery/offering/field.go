package offering

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func buildField(spec fieldSpec) (api.FieldDefinition, error) {
	kind, typeErr := fieldType(spec)
	storage, storageErr := fieldStorage(spec)
	filters, filterErr := filterCapabilities(spec.Filters)
	if typeErr != nil || storageErr != nil || filterErr != nil {
		return api.FieldDefinition{}, ErrContract
	}
	capabilities, err := api.NewFieldCapabilities(kind, api.FieldCapabilitiesInput{
		Selectable: true, Filters: filters, Sortable: spec.Sortable,
		CursorCanonicalBytes: spec.Cursor, Searchable: spec.Search,
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
