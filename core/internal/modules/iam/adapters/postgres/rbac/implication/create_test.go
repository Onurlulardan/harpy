package implicationpostgres

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

func TestCreateRejectsInvalidInput(t *testing.T) {
	var storage *store
	if _, err := storage.Create(context.Background(), tenantScope{}, rbacdomain.ImplicationInput{}); err == nil {
		t.Fatal("create accepted invalid input")
	}
}
