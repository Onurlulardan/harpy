package action

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestUpdateRejectsInvalidInput(t *testing.T) {
	var commands *Commands
	if _, err := commands.Update(context.Background(), platformtx.TenantScope{}, rbacdomain.UpdateCatalog{}); err == nil {
		t.Fatal("update accepted invalid input")
	}
}
