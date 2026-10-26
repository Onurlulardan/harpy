package api

import (
	"encoding/json"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func projectSystemResourceRead(
	read systemresources.ReadRecord,
) (map[string]json.RawMessage, int64, string, error) {
	record, recordErr := read.Record()
	catalog, catalogErr := read.Catalog()
	if recordErr != nil || catalogErr != nil {
		return nil, 0, "", systemresources.ErrUnavailable
	}
	id, revision, err := record.Identity()
	if err != nil {
		return nil, 0, "", systemresources.ErrUnavailable
	}
	values, err := record.Values()
	if err != nil {
		return nil, 0, "", systemresources.ErrUnavailable
	}
	raw, err := values.Values()
	if err != nil {
		return nil, 0, "", systemresources.ErrUnavailable
	}
	relations, err := read.Relations()
	if err != nil {
		return nil, 0, "", systemresources.ErrUnavailable
	}
	names, err := systemResourceDescriptorNames(catalog)
	if err != nil {
		return nil, 0, "", err
	}
	result := make(map[string]json.RawMessage, len(raw))
	for fieldID, value := range raw {
		name, found := names[fieldID]
		if !found || systemResourceSecretName(name) {
			return nil, 0, "", systemresources.ErrUnavailable
		}
		result[name] = append(json.RawMessage(nil), value...)
	}
	result["id"], _ = json.Marshal(id.String())
	result["revision"], _ = json.Marshal(revision)
	if err := projectSystemResourceRelations(result, relations); err != nil {
		return nil, 0, "", err
	}
	return result, revision, id.String(), nil
}

func systemResourceDescriptorNames(
	catalog systemresources.FieldCatalog,
) (map[systemresources.ID]string, error) {
	descriptors, err := catalog.Descriptors()
	if err != nil {
		return nil, systemresources.ErrUnavailable
	}
	result := make(map[systemresources.ID]string, len(descriptors))
	for _, descriptor := range descriptors {
		id, idErr := descriptor.ID()
		name, nameErr := descriptor.PublicName()
		if idErr != nil || nameErr != nil || systemResourceSecretName(name) {
			return nil, systemresources.ErrUnavailable
		}
		result[id] = name
	}
	return result, nil
}

func systemResourceAllowedNames(
	catalog systemresources.FieldCatalog,
) (map[string]struct{}, error) {
	byID, err := systemResourceDescriptorNames(catalog)
	if err != nil {
		return nil, err
	}
	result := make(map[string]struct{}, len(byID))
	for _, name := range byID {
		result[name] = struct{}{}
	}
	return result, nil
}

func systemResourceSecretName(name string) bool {
	switch name {
	case "password", "password_hash", "email_normalized", "token_version",
		"authorization_version", "identifier_id":
		return true
	default:
		return false
	}
}
