package authentication

import (
	"context"
	"testing"
)

func TestStartReauthenticationFailsClosedOutsideSessionScope(t *testing.T) {
	challenge, err := (Authenticator{}).StartReauthentication(
		context.Background(), CredentialScope{}, BeginReauthenticationRequest{},
	)
	if err == nil || challenge.valid {
		t.Fatalf("challenge valid = %t, err = %v", challenge.valid, err)
	}
}
