package domain

import "testing"

func TestHighestRoleUsesClosedRolePrecedence(t *testing.T) {
	if got := HighestRole(RoleViewer, RoleAdmin, RoleEditor); got != RoleAdmin {
		t.Fatalf("HighestRole = %q", got)
	}
	if BoardRole("OWNER").Valid() || PrincipalKind("GROUP").Valid() {
		t.Fatal("unknown role or principal kind was accepted")
	}
}
