package authentication

import (
	"context"
	"testing"
)

func TestCompleteReauthenticationFailsClosedOutsideSessionScope(t *testing.T) {
	grant, err := (Authenticator{}).CompleteReauthentication(
		context.Background(), CredentialScope{}, ReauthenticateRequest{},
	)
	if err == nil || grant.valid {
		t.Fatalf("grant valid = %t, err = %v", grant.valid, err)
	}
}
