package authentication

import (
	"context"
	"testing"
)

func TestRejectInitialPasswordFailsClosedOutsideCredentialScope(t *testing.T) {
	if err := (InitialPasswordAuthenticator{}).Reject(
		context.Background(), CredentialScope{}, nil,
	); err == nil {
		t.Fatal("initial password rejection accepted an invalid scope")
	}
}
