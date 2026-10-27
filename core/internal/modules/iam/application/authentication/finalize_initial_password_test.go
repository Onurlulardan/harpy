package authentication

import (
	"context"
	"testing"
)

func TestFinalizeInitialPasswordFailsClosedAndDestroysNewSecret(t *testing.T) {
	secret := &trackedPasswordSecret{}
	grant, err := (InitialPasswordAuthenticator{}).Finalize(
		context.Background(), CredentialScope{}, nil, secret, DeviceEvidence{},
	)
	if err == nil || !secret.destroyed || grant.valid {
		t.Fatalf("finalize err=%v destroyed=%t", err, secret.destroyed)
	}
}
