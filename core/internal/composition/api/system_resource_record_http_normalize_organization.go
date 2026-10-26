package api

import (
	"encoding/json"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func normalizeSystemOrganizationCreate(
	admission systemresources.Admission,
	document map[string]json.RawMessage,
) error {
	if err := normalizeSystemOrganizationOwners(document); err != nil {
		return err
	}
	return defaultSystemOrganizationOwners(admission, document)
}

func defaultSystemOrganizationOwners(
	admission systemresources.Admission,
	document map[string]json.RawMessage,
) error {
	actorID, _, _, err := admission.Identity()
	if err != nil {
		return systemresources.ErrContract
	}
	raw, present := document["owner_ids"]
	if present {
		var owners []string
		if json.Unmarshal(raw, &owners) != nil {
			return systemresources.ErrContract
		}
		if len(owners) > 0 {
			return nil
		}
	}
	encoded, _ := json.Marshal([]string{actorID.String()})
	document["owner_ids"] = encoded
	return nil
}

func normalizeSystemOrganizationOwners(
	document map[string]json.RawMessage,
) error {
	singular, hasSingular := document["owner_id"]
	_, hasPlural := document["owner_ids"]
	if hasSingular && hasPlural {
		return systemresources.ErrContract
	}
	if !hasSingular {
		return nil
	}
	var value string
	if json.Unmarshal(singular, &value) != nil {
		return systemresources.ErrContract
	}
	if _, err := systemresources.ParseID(value); err != nil {
		return systemresources.ErrContract
	}
	encoded, _ := json.Marshal([]string{value})
	document["owner_ids"] = encoded
	delete(document, "owner_id")
	return nil
}
