package lifecycle

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestUpgradeRejectsIncompleteCommand(t *testing.T) {
	_, err := (&Application{}).Upgrade(context.Background(), platformtx.TenantScope{}, UpgradeRequest{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("upgrade error = %v", err)
	}
}
