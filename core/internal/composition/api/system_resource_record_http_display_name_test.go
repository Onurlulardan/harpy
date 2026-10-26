package api

import (
	"encoding/json"
	"testing"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func TestNormalizeSystemResourceDisplayNameProjectsVisibleName(t *testing.T) {
	displayID, err := systemresources.ParseID("842de2f7-a4a9-7605-b7de-3063f65aa092")
	if err != nil {
		t.Fatal(err)
	}
	for _, resource := range []systemresources.Resource{
		systemresources.Roles, systemresources.Organizations,
	} {
		document := map[string]json.RawMessage{
			"name":         json.RawMessage(`"Operations"`),
			"display_name": json.RawMessage(`"Different"`),
		}
		err = normalizeSystemResourceDisplayName(
			resource, document,
			map[string]systemresources.ID{"display_name": displayID},
			map[systemresources.ID]bool{displayID: true},
		)
		if err != nil || string(document["display_name"]) != `"Operations"` {
			t.Fatalf("resource=%s document=%v err=%v", resource, document, err)
		}
	}
}

func TestNormalizeSystemResourceDisplayNameSkipsMissingCatalog(t *testing.T) {
	document := map[string]json.RawMessage{"name": json.RawMessage(`"Operations"`)}
	err := normalizeSystemResourceDisplayName(
		systemresources.Roles, document,
		map[string]systemresources.ID{},
		map[systemresources.ID]bool{},
	)
	if err != nil || document["display_name"] != nil {
		t.Fatalf("document=%v err=%v", document, err)
	}
}

func TestNormalizeSystemResourceDisplayNameLeavesUsersUntouched(t *testing.T) {
	displayID, _ := systemresources.ParseID("842de2f7-a4a9-7605-b7de-3063f65aa092")
	document := map[string]json.RawMessage{"display_name": json.RawMessage(`"Ada"`)}
	if err := normalizeSystemResourceDisplayName(
		systemresources.Users, document,
		map[string]systemresources.ID{"display_name": displayID},
		map[systemresources.ID]bool{displayID: true},
	); err != nil || string(document["display_name"]) != `"Ada"` {
		t.Fatalf("document=%v err=%v", document, err)
	}
}

func TestNormalizeSystemResourceDisplayNameDerivesLegacyUserName(t *testing.T) {
	displayID, _ := systemresources.ParseID("842de2f7-a4a9-7605-b7de-3063f65aa092")
	document := map[string]json.RawMessage{
		"first_name": json.RawMessage(`"Ada"`),
		"last_name":  json.RawMessage(`"Lovelace"`),
	}
	err := normalizeSystemResourceDisplayName(
		systemresources.Users, document,
		map[string]systemresources.ID{"display_name": displayID},
		map[systemresources.ID]bool{displayID: true},
	)
	if err != nil || string(document["display_name"]) != `"Ada Lovelace"` {
		t.Fatalf("document=%v err=%v", document, err)
	}
}

func TestNormalizeSystemResourceDisplayNameDropsReadOnlyProjection(t *testing.T) {
	displayID, _ := systemresources.ParseID("842de2f7-a4a9-7605-b7de-3063f65aa092")
	document := map[string]json.RawMessage{
		"name":         json.RawMessage(`"Operations"`),
		"display_name": json.RawMessage(`""`),
	}
	err := normalizeSystemResourceDisplayName(
		systemresources.Roles, document,
		map[string]systemresources.ID{"display_name": displayID},
		map[systemresources.ID]bool{displayID: false},
	)
	if err != nil || document["display_name"] != nil {
		t.Fatalf("document=%v err=%v", document, err)
	}
}
