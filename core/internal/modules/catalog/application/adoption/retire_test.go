package adoption

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRetireRejectsEmptyRoute(t *testing.T) {
	err := (&Application{}).Retire(context.Background(), platformtx.TenantScope{}, "")
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("retire compatibility edge error = %v", err)
	}
}
