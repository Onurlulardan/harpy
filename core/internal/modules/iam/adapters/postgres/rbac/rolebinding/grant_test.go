package rolebindingpostgres

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

func TestGrantRejectsInvalidInput(t *testing.T) {
	var storage *store
	if _, err := storage.Grant(context.Background(), tenantScope{}, rbacdomain.BindingCommand{}); err == nil {
		t.Fatal("grant accepted invalid input")
	}
}
