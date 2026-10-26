package authentication

import (
	"context"
	"testing"
)

func TestClaimInitialPasswordFailsClosedAndDestroysTemporarySecret(t *testing.T) {
	secret := &trackedPasswordSecret{}
	challenge, err := (InitialPasswordAuthenticator{}).Claim(
		context.Background(), CredentialScope{}, secret,
	)
	if err == nil || !secret.destroyed || challenge.material != nil {
		t.Fatalf("claim err=%v destroyed=%t", err, secret.destroyed)
	}
}
