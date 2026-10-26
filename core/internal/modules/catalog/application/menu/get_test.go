package menu

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGetFailsClosedWithoutApplication(t *testing.T) {
	_, err := (*Application)(nil).Get(context.Background(), platformtx.TenantScope{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("get menu error = %v", err)
	}
}
