package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestListRejectsUnboundedPage(t *testing.T) {
	_, err := (&Application{}).ListInstallations(context.Background(), platformtx.TenantScope{}, PageRequest{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("list installations error = %v", err)
	}
}
