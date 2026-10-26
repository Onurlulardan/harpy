package iam

import (
	"testing"
)

func TestBuildInstallationPresentationCatalogs(t *testing.T) {
	operationID := [16]byte{0x01, 0x8f, 0x20, 0x00, 0x00, 0x00, 0x70, 0x00,
		0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
	catalogs, err := BuildInstallationPresentationCatalogs(operationID)
	if err != nil {
		t.Fatal(err)
	}
	user, err := catalogs.User.Presentation()
	if err != nil {
		t.Fatal(err)
	}
	organization, err := catalogs.Organization.Presentation()
	if err != nil {
		t.Fatal(err)
	}
	role, err := catalogs.Role.Presentation()
	if err != nil {
		t.Fatal(err)
	}
	userFields, userErr := user.Fields()
	organizationFields, organizationErr := organization.Fields()
	roleFields, roleErr := role.Fields()
	if userErr != nil || organizationErr != nil || roleErr != nil {
		t.Fatalf("catalog field access failed: %v/%v/%v",
			userErr, organizationErr, roleErr)
	}
	for name, count := range map[string]int{
		"user": len(userFields), "organization": len(organizationFields),
		"role": len(roleFields),
	} {
		if count < 2 {
			t.Fatalf("%s catalog fields=%d", name, count)
		}
	}
	descriptors, err := catalogs.User.Descriptors()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"display_name": false, "first_name": false,
		"last_name": false, "phone": false,
	}
	for _, descriptor := range descriptors {
		name, nameErr := descriptor.PublicName()
		if nameErr == nil {
			if _, required := want[name]; required {
				want[name] = true
			}
		}
	}
	for name, found := range want {
		if !found {
			t.Fatalf("installation user catalog lacks HTTP alias %q", name)
		}
	}
}
