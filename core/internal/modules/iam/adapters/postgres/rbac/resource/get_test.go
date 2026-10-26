package resourcepostgres

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

func TestGetRejectsInvalidInput(t *testing.T) {
	var storage *store
	if _, err := storage.Get(context.Background(), tenantScope{}, rbacdomain.GetQuery{}); err == nil {
		t.Fatal("get accepted invalid input")
	}
}
