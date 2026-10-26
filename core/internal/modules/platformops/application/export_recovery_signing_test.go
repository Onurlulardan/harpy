package application

import (
	"bytes"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestExportRecoverySigningRefusesTamperedPersistence(t *testing.T) {
	now := time.Date(2026, 7, 23, 3, 0, 0, 0, time.UTC)
	bundle := buildRecoveryBundleFixture(t, now)
	record := ports.SigningBundleRecord{EnvironmentID: bundle.EnvironmentID(),
		Generation: 1, SHA256: bundle.SHA256(), BundleBytes: bundle.CanonicalBytes(), CreatedAt: now}
	store := &recoveryStoreFixture{records: map[int64]ports.SigningBundleRecord{1: record}}
	coordinator := NewRecoveryCoordinator(store, recoveryClockFixture{now: now})
	exported, err := coordinator.ExportRecoverySigning(t.Context(), bundle.EnvironmentID(), 1)
	if err != nil || !bytes.Equal(exported.BundleBytes, bundle.CanonicalBytes()) {
		t.Fatalf("export signing bundle = %#v, err=%v", exported, err)
	}
	record.BundleBytes[5] ^= 1
	store.records[1] = record
	if _, err = coordinator.ExportRecoverySigning(t.Context(), bundle.EnvironmentID(), 1); err != api.ErrConflict {
		t.Fatalf("tampered persisted bundle error = %v", err)
	}
}
