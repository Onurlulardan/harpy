package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRetireRejectsMissingPackage(t *testing.T) {
	_, err := (&Application{}).Retire(context.Background(), platformtx.TenantScope{}, PackageCommand{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("retire error = %v", err)
	}
}
