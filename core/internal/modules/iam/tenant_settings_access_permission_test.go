package iam

import "testing"

func TestResolveTenantSettingsAccessPinsOrganizationReadUpdate(t *testing.T) {
	read, valid := resolveTenantSettingsAccess(
		"tenancy.settings.read", "tenancy.settings.get")
	if !valid || read.resource != "iam.organizations" || read.action != "read" {
		t.Fatalf("read target = %#v/%t", read, valid)
	}
	for _, operation := range []string{
		"tenancy.settings.update", "tenancy.settings.test_email",
	} {
		target, ok := resolveTenantSettingsAccess("tenancy.settings.manage", operation)
		if !ok || target.resource != "iam.organizations" || target.action != "update" {
			t.Fatalf("manage target for %s = %#v/%t", operation, target, ok)
		}
	}
	for _, input := range [][2]string{
		{"tenancy.settings.manage", "tenancy.settings.get"},
		{"tenancy.settings.read", "tenancy.settings.update"},
		{"iam.organizations.update", "tenancy.settings.get"},
	} {
		if _, ok := resolveTenantSettingsAccess(input[0], input[1]); ok {
			t.Fatalf("tenant settings authority accepted %q/%q", input[0], input[1])
		}
	}
}
