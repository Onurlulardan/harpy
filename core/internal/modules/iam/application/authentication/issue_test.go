package authentication

import (
	"context"
	"testing"
)

type trackedPasswordSecret struct{ destroyed bool }

func (secret *trackedPasswordSecret) Destroy() { secret.destroyed = true }

func TestIssueFailsClosedAndDestroysSecretOutsideCredentialScope(t *testing.T) {
	secret := &trackedPasswordSecret{}
	grant, err := (Authenticator{}).Issue(
		context.Background(), CredentialScope{}, IssueRequest{Secret: secret},
	)
	if err == nil || grant.valid || !secret.destroyed {
		t.Fatalf("grant valid = %t, destroyed = %t, err = %v",
			grant.valid, secret.destroyed, err)
	}
}
