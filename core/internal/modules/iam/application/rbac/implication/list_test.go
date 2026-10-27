package implication

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestListRejectsInvalidInput(t *testing.T) {
	var commands *Commands
	if _, err := commands.List(context.Background(), platformtx.TenantScope{}, rbacdomain.ListQuery{}); err == nil {
		t.Fatal("list accepted invalid input")
	}
}
