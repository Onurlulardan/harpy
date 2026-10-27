package identitychange

import (
	"context"
	"testing"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestWriteAuditRejectsInvalidChangeBeforeBinding(t *testing.T) {
	var storage *store
	_, err := storage.WriteAudit(
		context.Background(), platformtx.TenantScope{}, identitydomain.Change{},
	)
	if err == nil {
		t.Fatal("invalid identity audit change was accepted")
	}
}
