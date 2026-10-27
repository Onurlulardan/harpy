package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func TestRejectInitialPasswordRejectsUnboundStore(t *testing.T) {
	if _, err := (*store)(nil).RejectInitialPassword(
		context.Background(), authentication.CredentialScope{},
		authentication.InitialPasswordReject{},
	); err == nil {
		t.Fatal("initial password rejection accepted an unbound store")
	}
}
