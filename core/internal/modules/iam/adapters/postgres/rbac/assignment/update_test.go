package assignmentpostgres

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

func TestUpdateRejectsInvalidAssignment(t *testing.T) {
	var storage *store
	if _, err := storage.Update(
		context.Background(), tenantScope{}, rbacdomain.AssignmentMutation{},
	); err == nil {
		t.Fatal("update accepted an invalid assignment")
	}
}
