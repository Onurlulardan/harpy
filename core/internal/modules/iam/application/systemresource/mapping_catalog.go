package systemresource

import publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"

func visibleFieldSet(
	catalog publicresource.FieldCatalog,
) (map[publicresource.ID]struct{}, error) {
	presentation, err := catalog.Presentation()
	if err != nil {
		return nil, publicresource.ErrContract
	}
	fields, err := presentation.Fields()
	if err != nil || len(fields) == 0 {
		return nil, publicresource.ErrContract
	}
	result := make(map[publicresource.ID]struct{}, len(fields))
	for _, field := range fields {
		text, idErr := field.ID()
		fieldID, parseErr := publicresource.ParseID(text)
		if idErr != nil || parseErr != nil {
			return nil, publicresource.ErrContract
		}
		result[fieldID] = struct{}{}
	}
	return result, nil
}

func fieldVisible(fields map[publicresource.ID]struct{}, fieldID string) bool {
	parsed, err := publicresource.ParseID(fieldID)
	if err != nil {
		return false
	}
	_, visible := fields[parsed]
	return visible
}
