package systemresourcequery

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRunCatalogRejectsMissingRuntime(t *testing.T) {
	var service *Service
	if _, err := service.RunCatalog(
		context.Background(), transaction.TenantScope{}, catalog.Published{}, Request{},
	); !errors.Is(err, ErrContract) {
		t.Fatalf("RunCatalog() error=%v", err)
	}
}
