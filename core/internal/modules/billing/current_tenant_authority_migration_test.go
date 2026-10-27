package billing

import (
	"strings"
	"testing"
)

func TestCurrentTenantAuthorityResolverRemainsBillingOwnedAndPrivate(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	var migration string
	for _, document := range documents {
		if strings.Contains(document.Path(), "/fresh/20260717010315_") {
			migration = string(document.Content())
			break
		}
	}
	if migration == "" {
		t.Fatal("billing current tenant authority resolver migration is missing")
	}
	for _, token := range []string{
		"platform.billing_resolve_current_tenant_authority_v1",
		"platform.billing_tenant_authorities", "platform.cutover_groups",
		"billing_member.domain = 'billing' AND billing_member.ordinal = 1",
		"group_row.state = 'v2' AND group_row.writer = 'v2'",
		"platform.billing_assert_tenant_authority_v1(v_authority_id)",
		"REVOKE ALL ON FUNCTION",
	} {
		if !strings.Contains(migration, token) {
			t.Fatalf("billing resolver migration misses %q", token)
		}
	}
	if strings.Contains(migration, "GRANT EXECUTE") {
		t.Fatal("billing resolver must not be executable directly by runtime")
	}
}

func TestCurrentTenantAuthorityReceiptFreezesAllSchemaSubmissionGenerations(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	var migration string
	for _, document := range documents {
		if strings.Contains(document.Path(), "/fresh/20260717010316_") {
			migration = string(document.Content())
			break
		}
	}
	if migration == "" {
		t.Fatal("billing authority receipt migration is missing")
	}
	for _, token := range []string{
		"platform.billing_resolve_current_tenant_authority_receipt_v1",
		"authority_id uuid", "authority_group_id text",
		"authority_group_generation bigint", "billing_generation bigint",
		"platform.billing_resolve_current_tenant_authority_v1(p_tenant_id)",
		"FOR SHARE", "REVOKE ALL ON FUNCTION",
	} {
		if !strings.Contains(migration, token) {
			t.Fatalf("billing authority receipt misses %q", token)
		}
	}
	if strings.Contains(migration, "GRANT EXECUTE") {
		t.Fatal("billing authority receipt must remain private")
	}
}
