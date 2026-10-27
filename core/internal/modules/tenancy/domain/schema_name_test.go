package domain

import (
	"regexp"
	"testing"
)

func TestSchemaNameFromUUIDUsesCanonicalBase32(t *testing.T) {
	name, err := SchemaNameFromUUID("00112233-4455-6677-8899-aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}
	const want = "t_aaisem2ekvthpcezvk54zxpo74"
	if name.String() != want {
		t.Fatalf("schema name = %q, want %q", name.String(), want)
	}
}

func TestSchemaNameIsAlwaysCanonicalIdentifier(t *testing.T) {
	pattern := regexp.MustCompile(`^t_[a-z2-7]{26}$`)
	inputs := []string{
		"00000000-0000-4000-8000-000000000000",
		"550e8400-e29b-41d4-a716-446655440000",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
	}
	for _, input := range inputs {
		name, err := SchemaNameFromUUID(input)
		if err != nil {
			t.Fatal(err)
		}
		if len(name.String()) != TenantSchemaNameLength || !pattern.MatchString(name.String()) {
			t.Fatalf("non-canonical schema name for %s: %q", input, name.String())
		}
	}
}

func TestSchemaNameDoesNotUseSlug(t *testing.T) {
	tenantID, err := ParseTenantID("550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatal(err)
	}
	first, err := SchemaNameForTenant(tenantID)
	if err != nil {
		t.Fatal(err)
	}
	second, err := SchemaNameForTenant(tenantID)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("schema identity is not deterministic: %q != %q", first, second)
	}
}

func TestSchemaNameRejectsZeroTenantID(t *testing.T) {
	if _, err := SchemaNameForTenant(TenantID{}); err == nil {
		t.Fatal("zero tenant ID generated a schema name")
	}
}
