package dashboard

import (
	"context"
	"errors"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"testing"
)

func TestListRejectsInvalidLimit(t *testing.T) {
	_, err := (&Application{}).List(context.Background(), platformtx.TenantScope{}, 0, nil)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("list error = %v", err)
	}
}
