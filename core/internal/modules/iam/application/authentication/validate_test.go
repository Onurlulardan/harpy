package authentication

import (
	"context"
	"testing"
)

func TestValidateRefreshFailsClosedAndDestroysCredentialOutsideScope(t *testing.T) {
	presented := &trackedPresentedRefresh{}
	err := (Authenticator{}).ValidateRefresh(
		context.Background(), CredentialScope{}, presented,
	)
	if err == nil || !presented.destroyed {
		t.Fatalf("destroyed = %t, err = %v", presented.destroyed, err)
	}
}

func TestBrowserClientComesOnlyFromAuthoritativeAccessMode(t *testing.T) {
	if browserClientForMode(AccessModeOperator) != ClientOperatorConsole ||
		browserClientForMode(AccessModePortal) != ClientPortal ||
		browserClientForMode("unknown") != "" {
		t.Fatal("browser client mapping is not fail-closed")
	}
}
