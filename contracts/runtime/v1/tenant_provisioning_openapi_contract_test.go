package runtimev1

import (
	"path/filepath"
	"testing"
)

func TestTenantProvisioningOpenAPIExposesExactPhase14Routes(t *testing.T) {
	directory := filepath.Join("..", "..", "http")
	core := readOpenAPIDocument(t,
		filepath.Join(directory, "core-api-paths", "platform.json"))
	expected := map[string]string{
		"/api/platform/tenants": "tenant-provisioning-v1.paths.json#/pathItems/tenants",
		"/api/admin/tenants":    "tenant-provisioning-v1.paths.json#/pathItems/legacyAdminTenants",
		"/api/platform/tenant-provisioning-operations/{operation_id}": "tenant-provisioning-v1.paths.json#/pathItems/operation",
		"/api/platform/tenants/{tenant_id}/admin-setup/reissue":       "tenant-provisioning-v1.paths.json#/pathItems/reissue",
	}
	for path, reference := range expected {
		item, ok := core[path].(map[string]any)
		if !ok || item["$ref"] != reference {
			t.Fatalf("path %s reference is invalid: %v", path, item)
		}
	}
	document := readOpenAPIDocument(t,
		filepath.Join(directory, "tenant-provisioning-v1.paths.json"))
	items, ok := document["pathItems"].(map[string]any)
	if !ok {
		t.Fatal("tenant provisioning path items are missing")
	}
	assertTenantProvisioningCapability(t, items, "tenants", "post",
		"platform.tenants.create", true)
	assertTenantProvisioningCapability(t, items, "legacyAdminTenants", "post",
		"platform.tenants.create", true)
	legacy := items["legacyAdminTenants"].(map[string]any)["post"].(map[string]any)
	if legacy["deprecated"] != true ||
		legacy["x-harpy-compat-policy"] != "exact-canonical-command-no-dual-writer" {
		t.Fatal("legacy tenant create route is not pinned to the canonical command")
	}
	assertTenantProvisioningCapability(t, items, "operation", "get",
		"platform.tenants.read", false)
	assertTenantProvisioningCapability(t, items, "reissue", "post",
		"platform.tenants.admin_setup.reissue", true)
	reissue := items["reissue"].(map[string]any)["post"].(map[string]any)
	if reissue["requestBody"] != nil {
		t.Fatal("initial admin setup reissue must remain bodyless")
	}
	assertTenantProvisioningOperationSchema(t, directory)
}

func assertTenantProvisioningOperationSchema(t *testing.T, directory string) {
	t.Helper()
	document := readOpenAPIDocument(t,
		filepath.Join(directory, "tenant-provisioning-v1.schemas.json"))
	definitions, ok := document["$defs"].(map[string]any)
	operation, operationOK := definitions["Operation"].(map[string]any)
	properties, propertiesOK := operation["properties"].(map[string]any)
	delivery, deliveryOK := properties["admin_setup_delivery_state"].(map[string]any)
	states, statesOK := delivery["enum"].([]any)
	branches, branchesOK := operation["oneOf"].([]any)
	if !ok || !operationOK || !propertiesOK || !deliveryOK || !statesOK ||
		!branchesOK || len(states) != 7 || len(branches) != 3 {
		t.Fatalf("tenant provisioning operation schema is not exact: %v", operation)
	}
	wantStates := map[string]bool{"pending": true, "delivered": true,
		"outcome_unknown": true, "permanent_failure": true, "accepted": true,
		"reissued": true, "revoked": true}
	for _, state := range states {
		value, valueOK := state.(string)
		if !valueOK || !wantStates[value] {
			t.Fatalf("unexpected admin setup delivery state: %v", state)
		}
		delete(wantStates, value)
	}
	if len(wantStates) != 0 {
		t.Fatalf("missing admin setup delivery states: %v", wantStates)
	}
}

func assertTenantProvisioningCapability(
	t *testing.T,
	items map[string]any,
	itemName, method, capability string,
	wantsIdempotency bool,
) {
	t.Helper()
	item, itemOK := items[itemName].(map[string]any)
	operation, operationOK := item[method].(map[string]any)
	if !itemOK || !operationOK ||
		operation["x-harpy-capability"] != capability {
		t.Fatalf("%s %s capability is invalid: %v", method, itemName, operation)
	}
	_, hasParameters := operation["parameters"]
	if hasParameters != wantsIdempotency {
		t.Fatalf("%s %s idempotency contract is invalid", method, itemName)
	}
}
