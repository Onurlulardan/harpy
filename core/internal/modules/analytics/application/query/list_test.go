package query

import (
	"context"
	"errors"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"testing"
)

func TestListRejectsInvalidPageBeforeScope(t *testing.T) {
	_, err := (&Application{}).List(context.Background(), platformtx.TenantScope{}, 101, nil)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("list error = %v", err)
	}
}
