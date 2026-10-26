package catalog

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestBackupVerifierRejectsMalformedBinding(t *testing.T) {
	_, err := NewBackupVerifier().Verify(context.Background(), platformtx.TenantScope{},
		"not-a-receipt", catalogapp.BackupRequirement{})
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatalf("verify error = %v", err)
	}
}
