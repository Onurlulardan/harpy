package dashboard

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReadDefaultRejectsUnboundScope(t *testing.T) {
	_, err := (&Application{}).ReadDefault(
		context.Background(), platformtx.TenantScope{}, "home",
	)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("read default error = %v", err)
	}
}
