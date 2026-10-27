package systemresourcequery

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestResolveCatalogRejectsMissingRuntime(t *testing.T) {
	var service *Service
	if _, err := service.ResolveCatalog(
		context.Background(), transaction.TenantScope{}, "users",
	); !errors.Is(err, ErrContract) {
		t.Fatalf("ResolveCatalog() error=%v", err)
	}
}
