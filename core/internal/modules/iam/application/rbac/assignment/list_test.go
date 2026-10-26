package assignment

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestListRejectsInvalidAssignmentQuery(t *testing.T) {
	var commands *Commands
	if _, err := commands.List(
		context.Background(), platformtx.TenantScope{}, rbacdomain.AssignmentListQuery{},
	); err == nil {
		t.Fatal("list accepted an invalid assignment query")
	}
}
