package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRetryRejectsMissingOperation(t *testing.T) {
	_, err := (&Application{}).Retry(context.Background(), platformtx.TenantScope{}, RetryRequest{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("retry error = %v", err)
	}
}
