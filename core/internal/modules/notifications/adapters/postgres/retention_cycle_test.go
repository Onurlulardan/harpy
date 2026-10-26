package postgres

import (
	"testing"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/retention"
)

func TestRetentionCycleRejectsInvalidClaim(t *testing.T) {
	if _, err := (&RetentionStore{}).ApplyCycle(t.Context(), retentionapp.Cycle{}); err != retentionapp.ErrClaimInvalid {
		t.Fatalf("ApplyCycle() error = %v", err)
	}
}
