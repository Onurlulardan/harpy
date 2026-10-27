package iam

import (
	"encoding/json"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

var profileSelfWritableFields = map[string]struct{}{
	"first_name": {}, "last_name": {}, "phone": {}, "avatar": {},
}

func profileSelfPatch(
	schema profileapp.Schema,
	values map[string]json.RawMessage,
) ([]byte, error) {
	if schema.Resource() != profileapp.ResourceUser ||
		schema.State() != profileapp.SchemaActive || values == nil ||
		len(values) == 0 || len(values) > len(profileSelfWritableFields) {
		return nil, ErrProfileHTTPInvalid
	}
	fields := make(map[string]profileapp.Field, len(schema.Fields()))
	for _, field := range schema.Fields() {
		fields[field.Key()] = field
	}
	patch := make(map[string]json.RawMessage, len(values))
	for key, value := range values {
		if _, allowed := profileSelfWritableFields[key]; !allowed {
			return nil, ErrProfileHTTPFieldForbidden
		}
		field, found := fields[key]
		if !found || field.State() != profileapp.FieldActive ||
			field.Classification() == profileapp.ClassificationRestricted ||
			len(value) == 0 {
			return nil, ErrProfileHTTPFieldForbidden
		}
		patch[field.ID().String()] = value
	}
	encoded, err := json.Marshal(patch)
	if err != nil {
		return nil, ErrProfileHTTPInvalid
	}
	return encoded, nil
}
