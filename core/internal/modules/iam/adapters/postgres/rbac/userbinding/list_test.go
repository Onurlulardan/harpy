package userbindingpostgres

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

func TestListRejectsInvalidInput(t *testing.T) {
	var storage *store
	if _, err := storage.List(context.Background(), tenantScope{}, rbacdomain.BindingListQuery{}); err == nil {
		t.Fatal("list accepted invalid input")
	}
}
