package assignmentpostgres

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

func TestGetRejectsInvalidAssignmentQuery(t *testing.T) {
	var storage *store
	if _, err := storage.Get(
		context.Background(), tenantScope{}, rbacdomain.AssignmentGetQuery{},
	); err == nil {
		t.Fatal("get accepted an invalid assignment query")
	}
}
