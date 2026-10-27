package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestDisableRejectsMissingPackage(t *testing.T) {
	_, err := (&Application{}).Disable(context.Background(), platformtx.TenantScope{}, PackageCommand{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("disable error = %v", err)
	}
}
