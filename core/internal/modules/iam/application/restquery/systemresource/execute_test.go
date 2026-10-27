package systemresourcequery

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestExecuteRejectsMissingRuntime(t *testing.T) {
	var service *Service
	_, err := service.Execute(
		context.Background(), transaction.TenantScope{}, "users", Request{},
	)
	if !errors.Is(err, ErrContract) {
		t.Fatalf("Execute() error=%v", err)
	}
}
