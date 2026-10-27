package iam

import (
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

type identitySystemResourceValues struct {
	resource publicresource.Resource
	values   map[publicresource.ID]json.RawMessage
}

func newIdentitySystemResourceValues(
	resource publicresource.Resource,
	values publicresource.FieldValues,
	catalog publicresource.FieldCatalog,
) (identitySystemResourceValues, error) {
	if values.Validate() != nil || catalog.Validate() != nil {
		return identitySystemResourceValues{}, publicresource.ErrContract
	}
	catalogResource, catalogErr := catalog.Resource()
	presentation, presentationErr := catalog.Presentation()
	raw, valuesErr := values.Values()
	if catalogErr != nil || presentationErr != nil || valuesErr != nil ||
		catalogResource != resource {
		return identitySystemResourceValues{}, publicresource.ErrContract
	}
	for id := range raw {
		if _, found, err := presentation.Field(id.String()); err != nil || !found {
			return identitySystemResourceValues{}, publicresource.ErrForbidden
		}
		editable, err := catalog.Editable(id)
		if err != nil || !editable {
			return identitySystemResourceValues{}, publicresource.ErrForbidden
		}
	}
	return identitySystemResourceValues{resource: resource, values: raw}, nil
}

func (values identitySystemResourceValues) raw(fieldID string) (json.RawMessage, bool) {
	id, err := publicresource.ParseID(fieldID)
	if err != nil {
		return nil, false
	}
	value, found := values.values[id]
	return append(json.RawMessage(nil), value...), found
}

func (values identitySystemResourceValues) text(
	fieldID string,
	required bool,
) (string, bool, error) {
	raw, found := values.raw(fieldID)
	if !found {
		if required {
			return "", false, publicresource.ErrContract
		}
		return "", false, nil
	}
	var result string
	if json.Unmarshal(raw, &result) != nil || required && result == "" {
		return "", false, publicresource.ErrContract
	}
	return result, true, nil
}

func (values identitySystemResourceValues) boolean(
	fieldID string,
) (bool, bool, error) {
	raw, found := values.raw(fieldID)
	if !found {
		return false, false, nil
	}
	var result bool
	if json.Unmarshal(raw, &result) != nil {
		return false, false, publicresource.ErrContract
	}
	return result, true, nil
}
