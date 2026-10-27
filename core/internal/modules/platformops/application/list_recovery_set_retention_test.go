package application

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestListRecoverySetRetention(t *testing.T) {
	store := &recoveryControlFixture{sets: map[string]ports.RecoverySetRecord{},
		restores: map[string]ports.RestoreRecord{}, retention: []ports.RetentionRecord{{
			RecoverySetID: "set-id", ManifestSHA256: appDigest("a"),
			ProtectedAt: time.Now().UTC(),
		}}}
	service := NewRecoveryService(store, store, store, store, recoveryClockFixture{})
	result, err := service.ListRecoverySetRetention(context.Background(), appUUID(1))
	if err != nil || len(result) != 1 || result[0].RecoverySetID != "set-id" {
		t.Fatalf("retention = %#v %v", result, err)
	}
}
