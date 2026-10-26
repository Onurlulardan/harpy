package postgres

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestWriteDecisionAuditRejectsUnownedScope(t *testing.T) {
	store, err := NewPermissionRequestDecisionAuditStore()
	if err != nil {
		t.Fatal(err)
	}
	if err := store.WriteDecisionAudit(
		context.Background(), platformtx.TenantScope{}, permissionrequest.DecisionRecord{},
	); err == nil {
		t.Fatal("audit store accepted an unowned scope")
	}
}
