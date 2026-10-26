package invitationaudit

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestPersistRejectsInvalidChangeBeforeBinding(t *testing.T) {
	var store *store
	if err := store.PersistTenant(
		context.Background(), platformtx.TenantScope{}, invitation.AuditChange{},
	); err == nil {
		t.Fatal("invalid invitation audit was accepted")
	}
}
