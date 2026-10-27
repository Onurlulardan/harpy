package assignmentpostgres

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

func TestRetireRejectsInvalidAssignment(t *testing.T) {
	var storage *store
	if _, err := storage.Retire(
		context.Background(), tenantScope{}, rbacdomain.AssignmentMutation{},
	); err == nil {
		t.Fatal("retire accepted an invalid assignment")
	}
}
