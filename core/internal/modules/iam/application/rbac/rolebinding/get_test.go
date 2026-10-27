package rolebinding

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGetRejectsInvalidInput(t *testing.T) {
	var commands *Commands
	if _, err := commands.Get(context.Background(), platformtx.TenantScope{}, rbacdomain.GetQuery{}); err == nil {
		t.Fatal("get accepted invalid input")
	}
}
