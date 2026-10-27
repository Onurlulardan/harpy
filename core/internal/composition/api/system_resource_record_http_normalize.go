package api

import (
	"bytes"
	"encoding/json"
	"strings"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func normalizeSystemResourceCreate(
	admission systemresources.Admission,
	resource systemresources.Resource,
	document map[string]json.RawMessage,
) error {
	normalizeSystemResourceNullableRelations(resource, document)
	switch resource {
	case systemresources.Users:
		return normalizeSystemUserCreate(document)
	case systemresources.Roles:
		return normalizeSystemRoleCreate(admission, document)
	case systemresources.Organizations:
		return normalizeSystemOrganizationCreate(admission, document)
	default:
		return systemresources.ErrContract
	}
}

func normalizeSystemUserCreate(document map[string]json.RawMessage) error {
	singular, hasSingular := document["organization_id"]
	_, hasPlural := document["organization_ids"]
	if hasSingular && hasPlural {
		return systemresources.ErrContract
	}
	if hasSingular {
		var id string
		if json.Unmarshal(singular, &id) != nil || id == "" {
			return systemresources.ErrContract
		}
		encoded, _ := json.Marshal([]string{id})
		document["organization_ids"] = encoded
		delete(document, "organization_id")
	}
	if _, present := document["status"]; present {
		delete(document, "status")
	}
	return nil
}

func normalizeSystemRoleCreate(
	admission systemresources.Admission,
	document map[string]json.RawMessage,
) error {
	_, tenantID, _, err := admission.Identity()
	if err != nil {
		return systemresources.ErrContract
	}
	kind, err := systemResourceOptionalText(document["scope_kind"])
	if err != nil {
		return err
	}
	kind = strings.ToLower(kind)
	organization, err := systemResourceOptionalText(document["organization_id"])
	if err != nil {
		return err
	}
	scope, err := systemResourceOptionalText(document["scope_id"])
	if err != nil {
		return err
	}
	if kind == "" {
		if organization != "" {
			kind = "organization"
		} else {
			kind = "tenant"
		}
	}
	switch kind {
	case "tenant":
		scope = tenantID.String()
		delete(document, "organization_id")
	case "organization":
		if organization == "" {
			organization = scope
		}
		if organization == "" || scope != "" && scope != organization {
			return systemresources.ErrContract
		}
		scope = organization
		document["organization_id"], _ = json.Marshal(organization)
	default:
		return systemresources.ErrContract
	}
	document["scope_kind"], _ = json.Marshal(kind)
	document["scope_id"], _ = json.Marshal(scope)
	normalizeSystemRoleDefault(document)
	return nil
}

func normalizeSystemRoleDefault(document map[string]json.RawMessage) {
	raw, present := document["is_default"]
	trimmed := bytes.TrimSpace(raw)
	if !present || len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) ||
		bytes.Equal(trimmed, []byte(`""`)) {
		document["is_default"], _ = json.Marshal(false)
		return
	}
	var flag bool
	if json.Unmarshal(trimmed, &flag) == nil {
		document["is_default"], _ = json.Marshal(flag)
		return
	}
	var text string
	if json.Unmarshal(trimmed, &text) != nil {
		return
	}
	if strings.EqualFold(text, "true") || strings.EqualFold(text, "false") {
		document["is_default"], _ = json.Marshal(strings.EqualFold(text, "true"))
	}
}

func normalizeSystemResourceNullableRelations(
	resource systemresources.Resource,
	document map[string]json.RawMessage,
) {
	names := map[systemresources.Resource][]string{
		systemresources.Roles:         {"scope_id", "organization_id"},
		systemresources.Organizations: {"parent_id"},
	}[resource]
	for _, name := range names {
		raw, present := document[name]
		if present && bytes.Equal(bytes.TrimSpace(raw), []byte(`""`)) {
			document[name] = json.RawMessage("null")
		}
	}
}

func systemResourceOptionalText(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return "", nil
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return "", systemresources.ErrContract
	}
	return value, nil
}
