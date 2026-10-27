package protectedadminpostgres

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

func TestUpdateRejectsInvalidInput(t *testing.T) {
	var storage *store
	if _, err := storage.Update(context.Background(), tenantScope{}, rbacdomain.UpdatePolicyCommand{}); err == nil {
		t.Fatal("update accepted invalid input")
	}
}
