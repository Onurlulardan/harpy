package userbinding

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRevokeRejectsInvalidInput(t *testing.T) {
	var commands *Commands
	if _, err := commands.Revoke(context.Background(), platformtx.TenantScope{}, rbacdomain.BindingCommand{}); err == nil {
		t.Fatal("revoke accepted invalid input")
	}
}
