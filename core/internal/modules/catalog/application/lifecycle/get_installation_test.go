package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestGetRejectsMissingPackage(t *testing.T) {
	_, err := (&Application{}).GetInstallation(context.Background(), platformtx.TenantScope{}, uuid.Nil)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("get installation error = %v", err)
	}
}
