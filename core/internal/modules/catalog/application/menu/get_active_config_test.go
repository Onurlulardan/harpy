package menu

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGetActiveConfigFailsClosedWithoutApplication(t *testing.T) {
	document, err := (*Application)(nil).GetActiveConfig(context.Background(), platformtx.TenantScope{})
	if document != nil || !errors.Is(err, ErrUnavailable) {
		t.Fatalf("get active menu config error = %v document=%v", err, document)
	}
}
