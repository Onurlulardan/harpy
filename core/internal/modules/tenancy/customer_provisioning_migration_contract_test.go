package tenancy

import (
	"strings"
	"testing"
)

func TestCustomerProvisioningMigrationsLockActiveOwnerlessReissueContract(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	selected := map[string]string{}
	for _, document := range documents {
		if !strings.Contains(document.Path(), "/platform/fresh/") {
			continue
		}
		for _, version := range []string{"20260722010007", "20260722080200", "20260722120010"} {
			if strings.Contains(document.Path(), version) {
				selected[version] = string(document.Content())
			}
		}
	}
	storage := selected["20260722010007"]
	activation := selected["20260722080200"]
	reissue := selected["20260722120010"]
	for _, token := range []string{
		"operation_kind = 'saas_customer_tenant_provision'",
		"binding_purpose = 'saas_customer'", "instance_owner_binding_id IS NULL",
	} {
		if !strings.Contains(storage, token) {
			t.Fatalf("ownerless customer storage lacks %q", token)
		}
	}
	if !strings.Contains(activation, "p_tenant_slug, 'ACTIVE'") {
		t.Fatal("customer activation no longer starts with ACTIVE tenant status")
	}
	for _, token := range []string{
		"'permanent_failure','outcome_unknown'", "setup_delivery_state='reissued'",
		"setup_reissue_operation_id=p_reissue_operation_id",
		"p_old_artifact_id=p_new_artifact_id", "harpy-v2:core-api",
	} {
		if !strings.Contains(reissue, token) {
			t.Fatalf("customer setup reissue migration lacks %q", token)
		}
	}
}
