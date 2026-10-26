package rbacchange

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestWriteAuditRejectsInvalidChangeBeforeBinding(t *testing.T) {
	var storage *store
	if err := storage.WriteAudit(context.Background(), platformtx.TenantScope{}, rbacdomain.Change{}); err == nil {
		t.Fatal("invalid audit change was accepted")
	}
}
