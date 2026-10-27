package catalog

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestAsyncAbortRejectsEmptyOperation(t *testing.T) {
	err := (&AsyncPort{}).Abort(context.Background(), platformtx.TenantScope{}, uuid.Nil)
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatalf("abort operation job error = %v", err)
	}
}
