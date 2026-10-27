package profilebackfill

import (
	"context"
	"errors"
	"testing"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func TestRecoveryInventoryRejectsUnboundScope(t *testing.T) {
	_, err := (*RecoveryReader)(nil).OpenInventory(
		context.Background(), tenantScope{},
		uuid.MustParse("12000000-0000-4000-8000-000000000001"),
		profilebackfillapp.Evidence{},
	)
	if !errors.Is(err, profilebackfillapp.ErrInvalid) {
		t.Fatalf("OpenInventory() error = %v", err)
	}
}
