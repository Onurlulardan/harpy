package api

import (
	"encoding/json"
	"strings"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func normalizeSystemResourceDisplayName(
	resource systemresources.Resource,
	document map[string]json.RawMessage,
	byName map[string]systemresources.ID,
	editable map[systemresources.ID]bool,
) error {
	if resource == systemresources.Users {
		return normalizeSystemUserDisplayName(document, byName, editable)
	}
	if resource != systemresources.Roles && resource != systemresources.Organizations {
		return nil
	}
	displayID, catalogued := byName["display_name"]
	name, hasName := document["name"]
	if !hasName {
		delete(document, "display_name")
		return nil
	}
	var nameValue string
	if json.Unmarshal(name, &nameValue) != nil || nameValue == "" {
		return systemresources.ErrContract
	}
	if !catalogued || !editable[displayID] {
		delete(document, "display_name")
		return nil
	}
	document["display_name"] = append(json.RawMessage(nil), name...)
	return nil
}

func normalizeSystemUserDisplayName(
	document map[string]json.RawMessage,
	byName map[string]systemresources.ID,
	editable map[systemresources.ID]bool,
) error {
	first, hasFirst := document["first_name"]
	last, hasLast := document["last_name"]
	if !hasFirst || !hasLast {
		return nil
	}
	displayID, catalogued := byName["display_name"]
	if !catalogued || !editable[displayID] {
		delete(document, "display_name")
		return nil
	}
	var firstValue, lastValue string
	if json.Unmarshal(first, &firstValue) != nil ||
		json.Unmarshal(last, &lastValue) != nil ||
		firstValue == "" || lastValue == "" ||
		strings.TrimSpace(firstValue) != firstValue ||
		strings.TrimSpace(lastValue) != lastValue {
		return systemresources.ErrContract
	}
	derived, _ := json.Marshal(firstValue + " " + lastValue)
	document["display_name"] = derived
	return nil
}
