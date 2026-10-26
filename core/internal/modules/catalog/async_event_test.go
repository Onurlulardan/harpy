package catalog

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestAsyncEmitRejectsEmptyEvent(t *testing.T) {
	err := (&AsyncPort{}).Emit(context.Background(), platformtx.TenantScope{}, catalogapp.InstallationEvent{})
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatalf("emit installation event error = %v", err)
	}
}
