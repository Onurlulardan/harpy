package iam

import (
	"os"
	"strings"
	"testing"
)

func TestIdentityOrganizationAndRoleCommandsKeepNameProjectionWired(t *testing.T) {
	for _, contract := range []struct {
		path     string
		required []string
	}{
		{"identity_system_resource_create_organization.go",
			[]string{"port.createNameProfile(", "ProfileOrganization", "values, name"}},
		{"identity_system_resource_create_role.go",
			[]string{"port.createNameProfile(", "ProfileRole", "values, name"}},
		{"identity_system_resource_update_organization.go",
			[]string{"port.updateNameProfile(", "ProfileOrganization",
				"current.Profile.SchemaID()", "values, name"}},
		{"identity_system_resource_update_role.go",
			[]string{"port.updateNameProfile(", "ProfileRole",
				"current.Profile.SchemaID()", "values, name"}},
	} {
		content, err := os.ReadFile(contract.path)
		if err != nil {
			t.Fatalf("read %s: %v", contract.path, err)
		}
		for _, required := range contract.required {
			if !strings.Contains(string(content), required) {
				t.Errorf("%s lost %q", contract.path, required)
			}
		}
	}
}

func TestIdentityUserCommandsDoNotUseNameProjection(t *testing.T) {
	for _, path := range []string{
		"identity_system_resource_create_user.go",
		"identity_system_resource_update_user.go",
	} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(string(content), "NameProfile(") {
			t.Errorf("%s unexpectedly uses organization/role projection", path)
		}
	}
}
