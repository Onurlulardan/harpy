package organizationbindingpostgres

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

func TestRevokeRejectsInvalidInput(t *testing.T) {
	var storage *store
	if _, err := storage.Revoke(context.Background(), tenantScope{}, rbacdomain.BindingCommand{}); err == nil {
		t.Fatal("revoke accepted invalid input")
	}
}
