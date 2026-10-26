package menu

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestUpdateRejectsEmptyOverride(t *testing.T) {
	_, err := (&Application{}).UpdateOverride(
		context.Background(), platformtx.TenantScope{}, UpdateRequest{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("update menu override error = %v", err)
	}
}
