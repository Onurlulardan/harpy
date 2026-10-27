package resourcepostgres

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

func TestRetireRejectsInvalidInput(t *testing.T) {
	var storage *store
	if _, err := storage.Retire(context.Background(), tenantScope{}, rbacdomain.StateCommand{}); err == nil {
		t.Fatal("retire accepted invalid input")
	}
}
