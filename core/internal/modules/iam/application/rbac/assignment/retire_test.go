package assignment

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRetireRejectsInvalidAssignment(t *testing.T) {
	var commands *Commands
	if _, err := commands.Retire(
		context.Background(), platformtx.TenantScope{}, rbacdomain.AssignmentMutation{},
	); err == nil {
		t.Fatal("retire accepted an invalid assignment")
	}
}
