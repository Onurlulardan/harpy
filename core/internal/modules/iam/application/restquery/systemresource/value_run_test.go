package systemresourcequery

import (
	"context"
	"errors"
	"testing"

	restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRunValueRejectsMissingRuntime(t *testing.T) {
	var service *Service
	if _, err := service.RunValue(
		context.Background(), transaction.TenantScope{},
		catalog.Published{}, restquery.QueryRequest{},
	); !errors.Is(err, ErrContract) {
		t.Fatalf("RunValue() error=%v", err)
	}
}
