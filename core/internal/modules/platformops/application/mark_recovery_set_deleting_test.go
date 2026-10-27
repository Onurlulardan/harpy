package application

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestMarkRecoverySetDeleting(t *testing.T) {
	store := &recoveryControlFixture{sets: map[string]ports.RecoverySetRecord{},
		restores: map[string]ports.RestoreRecord{}, retention: []ports.RetentionRecord{{
			RecoverySetID: "set-id", ManifestSHA256: appDigest("a"),
			ProtectedAt: time.Now().UTC(),
		}}}
	service := NewRecoveryService(store, store, store, store,
		recoveryClockFixture{now: time.Now().UTC()})
	if err := service.MarkRecoverySetDeleting(
		context.Background(), "set-id", appDigest("a"),
	); err != nil {
		t.Fatal(err)
	}
}
