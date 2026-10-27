package catalog

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestAsyncDispatchRejectsEmptyOperation(t *testing.T) {
	err := (&AsyncPort{}).Dispatch(context.Background(), platformtx.TenantScope{}, catalogapp.Operation{})
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatalf("dispatch operation error = %v", err)
	}
}
