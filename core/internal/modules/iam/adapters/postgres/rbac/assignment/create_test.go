package assignmentpostgres

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

func TestCreateRejectsInvalidAssignment(t *testing.T) {
	var storage *store
	if _, err := storage.Create(
		context.Background(), tenantScope{}, rbacdomain.AssignmentMutation{},
	); err == nil {
		t.Fatal("create accepted an invalid assignment")
	}
}
