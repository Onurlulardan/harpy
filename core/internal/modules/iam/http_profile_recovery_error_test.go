package iam

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
)

func TestProfileRecoveryRequiredUsesStableHTTPContract(t *testing.T) {
	status, code, _ := profileHTTPFailure(
		profileApplicationError(profileapp.ErrRecoveryRequired),
	)
	if status != 409 || code != "PROFILE_RECOVERY_REQUIRED" {
		t.Fatalf("profile recovery failure = %d/%s", status, code)
	}
	if !errors.Is(
		identityApplicationError(profileapp.ErrRecoveryRequired),
		ErrIdentityHTTPRecoveryRequired,
	) {
		t.Fatal("identity recovery failure lost stable sentinel")
	}
	identity, found := identityFailure(identityFailureRecoveryRequired)
	if !found || identity.status != 409 || identity.code != "PROFILE_RECOVERY_REQUIRED" {
		t.Fatalf("identity recovery failure = %+v/%v", identity, found)
	}
	status, code, _ = invitationFailure(
		mapInvitationApplicationError(invitation.ErrRecoveryRequired),
	)
	if status != 409 || code != "PROFILE_RECOVERY_REQUIRED" {
		t.Fatalf("invitation recovery failure = %d/%s", status, code)
	}
}

func TestProfileRecoveryBudgetUsesStableHTTPContract(t *testing.T) {
	mapped := profileApplicationError(profilebackfillapp.ErrBudgetExceeded)
	status, code, _ := profileHTTPFailure(mapped)
	if status != 409 || code != "PROFILE_RECOVERY_BUDGET_EXCEEDED" {
		t.Fatalf("profile recovery budget failure = %d/%s", status, code)
	}
}
