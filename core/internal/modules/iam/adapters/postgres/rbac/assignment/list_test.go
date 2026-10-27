package assignmentpostgres

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

func TestListRejectsInvalidAssignmentQuery(t *testing.T) {
	var storage *store
	if _, err := storage.List(
		context.Background(), tenantScope{}, rbacdomain.AssignmentListQuery{},
	); err == nil {
		t.Fatal("list accepted an invalid assignment query")
	}
}
