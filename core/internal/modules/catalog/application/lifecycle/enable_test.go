package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestEnableRejectsMissingPackage(t *testing.T) {
	_, err := (&Application{}).Enable(context.Background(), platformtx.TenantScope{}, PackageCommand{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("enable error = %v", err)
	}
}
