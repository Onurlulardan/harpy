package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRestoreRejectsMissingPackage(t *testing.T) {
	_, err := (&Application{}).Restore(context.Background(), platformtx.TenantScope{}, PackageCommand{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("restore error = %v", err)
	}
}
