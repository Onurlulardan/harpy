package assignment

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGetRejectsInvalidAssignmentQuery(t *testing.T) {
	var commands *Commands
	if _, err := commands.Get(
		context.Background(), platformtx.TenantScope{}, rbacdomain.AssignmentGetQuery{},
	); err == nil {
		t.Fatal("get accepted an invalid assignment query")
	}
}
