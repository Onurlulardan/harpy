package profilebackfill

import (
	"context"
	"errors"
	"testing"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestRecoveryParentRejectsUnownedScope(t *testing.T) {
	reader := &RecoveryReader{value: &store{}}
	_, err := reader.InspectParent(
		context.Background(), platformtx.TenantScope{},
		uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		profilebackfillapp.Evidence{},
	)
	if !errors.Is(err, profilebackfillapp.ErrInvalid) {
		t.Fatalf("InspectParent() error = %v", err)
	}
}
