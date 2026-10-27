package profilebackfill

import (
	"context"
	"errors"
	"testing"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
)

func TestRecoveryReceiptRejectsUnboundScope(t *testing.T) {
	_, _, err := (*RecoveryReader)(nil).LookupReceipt(
		context.Background(), tenantScope{}, profilebackfillapp.Evidence{},
	)
	if !errors.Is(err, profilebackfillapp.ErrInvalid) {
		t.Fatalf("LookupReceipt() error = %v", err)
	}
}
