package authentication

import (
	"context"
	"testing"
)

func TestRevokeSessionFailsClosedAndDestroysCredentialOutsideScope(t *testing.T) {
	presented := &trackedPresentedRefresh{}
	err := (Authenticator{}).RevokeSession(
		context.Background(), CredentialScope{}, LogoutRequest{Presented: presented},
	)
	if err == nil || !presented.destroyed {
		t.Fatalf("destroyed = %t, err = %v", presented.destroyed, err)
	}
}
