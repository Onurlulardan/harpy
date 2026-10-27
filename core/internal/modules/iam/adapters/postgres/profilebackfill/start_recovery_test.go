package profilebackfill

import (
	"context"
	"errors"
	"testing"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
)

func TestStartRecoveryRejectsUnboundScope(t *testing.T) {
	_, err := (*store)(nil).StartRecovery(
		context.Background(), tenantScope{},
		profilebackfillapp.PreparedRecoveryPlan{},
	)
	if !errors.Is(err, profilebackfillapp.ErrInvalid) {
		t.Fatalf("StartRecovery() error = %v", err)
	}
}
