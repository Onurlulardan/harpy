package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestReadOperationRejectsMissingOperation(t *testing.T) {
	_, err := (&Application{}).ReadOperation(
		context.Background(), platformtx.TenantScope{}, uuid.Nil,
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("read operation error = %v", err)
	}
}
