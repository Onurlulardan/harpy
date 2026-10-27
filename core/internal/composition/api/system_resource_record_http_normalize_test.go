package api

import (
	"encoding/json"
	"errors"
	"testing"

	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func TestNormalizeSystemOrganizationOwnersAdaptsLegacyOwnerID(t *testing.T) {
	document := map[string]json.RawMessage{
		"owner_id": json.RawMessage(`"67000000-0000-4000-8000-000000000001"`),
	}
	if err := normalizeSystemOrganizationOwners(document); err != nil ||
		string(document["owner_ids"]) != `["67000000-0000-4000-8000-000000000001"]` ||
		document["owner_id"] != nil {
		t.Fatalf("document=%v err=%v", document, err)
	}
}

func TestNormalizeSystemOrganizationOwnersRejectsAmbiguousInput(t *testing.T) {
	document := map[string]json.RawMessage{
		"owner_id":  json.RawMessage(`"67000000-0000-4000-8000-000000000001"`),
		"owner_ids": json.RawMessage(`["67000000-0000-4000-8000-000000000001"]`),
	}
	if err := normalizeSystemOrganizationOwners(document); !errors.Is(err, systemresources.ErrContract) {
		t.Fatalf("expected contract failure, got %v", err)
	}
}

func TestNormalizeSystemRoleDefaultCoercesEmptyText(t *testing.T) {
	document := map[string]json.RawMessage{"is_default": json.RawMessage(`""`)}
	normalizeSystemRoleDefault(document)
	if string(document["is_default"]) != "false" {
		t.Fatalf("document=%v", document)
	}
}

func TestNormalizeSystemRoleCreateFillsTenantScopeAndDefault(t *testing.T) {
	tenant, _ := systemresources.ParseID(recordHTTPTestTenant)
	actorID, _ := systemresources.ParseID(recordHTTPTestActor)
	actor, _ := systemresources.NewActorEvidence(tenant, actorID, 1, 1)
	admission, _ := systemresources.NewAdmission(tenant, actor, 7, 7, 7)
	document := map[string]json.RawMessage{
		"name":            json.RawMessage(`"Operator"`),
		"description":     json.RawMessage(`"test"`),
		"scope_kind":      json.RawMessage(`"Tenant"`),
		"organization_id": json.RawMessage(`""`),
	}
	if err := normalizeSystemResourceCreate(
		admission, systemresources.Roles, document,
	); err != nil {
		t.Fatal(err)
	}
	if string(document["scope_kind"]) != `"tenant"` ||
		string(document["scope_id"]) != `"`+recordHTTPTestTenant+`"` ||
		string(document["is_default"]) != "false" ||
		document["organization_id"] != nil {
		t.Fatalf("document=%v", document)
	}
}

func TestNormalizeSystemRoleDefaultCoercesBooleanText(t *testing.T) {
	document := map[string]json.RawMessage{"is_default": json.RawMessage(`"true"`)}
	normalizeSystemRoleDefault(document)
	if string(document["is_default"]) != "true" {
		t.Fatalf("document=%v", document)
	}
}

func TestNormalizeSystemOrganizationCreateDefaultsActorOwner(t *testing.T) {
	tenant, _ := systemresources.ParseID(recordHTTPTestTenant)
	actorID, _ := systemresources.ParseID(recordHTTPTestActor)
	actor, _ := systemresources.NewActorEvidence(tenant, actorID, 1, 1)
	admission, _ := systemresources.NewAdmission(tenant, actor, 7, 7, 7)
	document := map[string]json.RawMessage{"name": json.RawMessage(`"Forge"`)}
	if err := normalizeSystemResourceCreate(
		admission, systemresources.Organizations, document,
	); err != nil {
		t.Fatal(err)
	}
	if string(document["owner_ids"]) != `["`+recordHTTPTestActor+`"]` {
		t.Fatalf("document=%v", document)
	}
}
