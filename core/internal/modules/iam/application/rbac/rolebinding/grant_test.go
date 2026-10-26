package rolebinding

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGrantRejectsInvalidInput(t *testing.T) {
	var commands *Commands
	if _, err := commands.Grant(context.Background(), platformtx.TenantScope{}, rbacdomain.BindingCommand{}); err == nil {
		t.Fatal("grant accepted invalid input")
	}
}
