package assignment

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestUpdateRejectsInvalidAssignment(t *testing.T) {
	var commands *Commands
	if _, err := commands.Update(
		context.Background(), platformtx.TenantScope{}, rbacdomain.AssignmentMutation{},
	); err == nil {
		t.Fatal("update accepted an invalid assignment")
	}
}
