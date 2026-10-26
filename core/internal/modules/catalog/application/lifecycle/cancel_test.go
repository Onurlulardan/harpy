package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCancelRejectsMissingOperation(t *testing.T) {
	_, err := (&Application{}).Cancel(context.Background(), platformtx.TenantScope{}, CancelRequest{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("cancel error = %v", err)
	}
}
