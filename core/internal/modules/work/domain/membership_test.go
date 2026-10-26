package domain

import "testing"

func TestMembershipProtectsLastActiveUserAdmin(t *testing.T) {
	user := mustID(t, "018f9f3e-7c16-7a21-8b42-116a9d3c9870")
	membershipID := mustID(t, "018f9f3e-7c16-7a21-8b42-116a9d3c9871")
	membership := Membership{ID: membershipID, BoardID: user,
		Principal: Principal{Kind: PrincipalUser, ID: user}, Role: RoleAdmin,
		Active: true, RowVersion: 1}
	if err := ValidateAdminRemoval([]Membership{membership}, membershipID); err == nil {
		t.Fatal("last active USER/ADMIN removal was accepted")
	}
	if role := EffectiveRole([]Membership{membership}, []Principal{membership.Principal}); role != RoleAdmin {
		t.Fatalf("effective role = %q", role)
	}
}
