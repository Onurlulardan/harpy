package postgresv1

import "testing"

func TestV1AdoptedRoleContractRequiresRecoveryPrincipals(t *testing.T) {
	manifest := validRolesContract(t, "v1_adopted")
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	manifest.Roles = append(manifest.Roles[:9], manifest.Roles[10:]...)
	if err := manifest.Validate(); err == nil {
		t.Fatal("V1-adopted manifest without backup reader was accepted")
	}
}

func TestV1AdoptedPrivilegesRequireRecoveryAuthority(t *testing.T) {
	manifest := validPrivilegesContract(t, "v1_adopted")
	if err := manifest.Validate(); err != nil {
		t.Fatal(err)
	}
	manifest.Memberships = append(manifest.Memberships[:1], manifest.Memberships[2:]...)
	if err := manifest.Validate(); err == nil {
		t.Fatal("V1-adopted privileges without backup SET membership were accepted")
	}
}
