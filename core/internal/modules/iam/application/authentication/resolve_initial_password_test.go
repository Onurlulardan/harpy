package authentication

import (
	"context"
	"testing"
)

func TestResolveInitialPasswordFailsClosedOutsideCredentialScope(t *testing.T) {
	verified, err := (InitialPasswordAuthenticator{}).Resolve(
		context.Background(), CredentialScope{}, nil,
	)
	if err == nil || verified.valid {
		t.Fatalf("resolve err=%v valid=%t", err, verified.valid)
	}
}
