package action

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCreateRejectsInvalidInput(t *testing.T) {
	var commands *Commands
	if _, err := commands.Create(context.Background(), platformtx.TenantScope{}, rbacdomain.ActionInput{}); err == nil {
		t.Fatal("create accepted invalid input")
	}
}
