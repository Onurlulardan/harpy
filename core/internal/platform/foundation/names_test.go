package foundation

import "testing"

func TestTenantSchemaNameIsCanonicalAndStable(t *testing.T) {
	if _, err := TenantSchemaName([16]byte{}); err == nil {
		t.Fatal("accepted the zero tenant ID")
	}
	if _, err := TenantRoleName([16]byte{}); err == nil {
		t.Fatal("accepted the zero tenant ID")
	}
	id := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	firstSchema, err := TenantSchemaName(id)
	if err != nil {
		t.Fatal(err)
	}
	secondSchema, err := TenantSchemaName(id)
	if err != nil || firstSchema != secondSchema || !tenantSchemaPattern.MatchString(firstSchema) {
		t.Fatalf("schema = %s", firstSchema)
	}
	role, err := TenantRoleName(id)
	if err != nil || !tenantRolePattern.MatchString(role) || role == firstSchema {
		t.Fatalf("role = %s", role)
	}
	if !AllowedSchema(firstSchema) || AllowedSchema("tenant_user_input") || AllowedSchema("public;drop") {
		t.Fatal("schema allowlist drifted")
	}
	if AllowedExtension("postgis") || !AllowedExtension(VectorExtension) || !AllowedExtension("plpgsql") {
		t.Fatal("extension allowlist drifted")
	}
}
