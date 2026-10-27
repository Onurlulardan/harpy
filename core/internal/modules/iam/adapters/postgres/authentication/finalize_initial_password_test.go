package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func TestFinalizeInitialPasswordRejectsUnboundStore(t *testing.T) {
	if _, err := (*store)(nil).FinalizeInitialPassword(
		context.Background(), authentication.CredentialScope{},
		authentication.InitialPasswordFinalize{},
	); err == nil {
		t.Fatal("initial password finalization accepted an unbound store")
	}
}
