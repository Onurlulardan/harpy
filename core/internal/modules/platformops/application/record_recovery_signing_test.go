package application

import (
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestRecordRecoverySigningPersistsCanonicalGeneration(t *testing.T) {
	now := time.Date(2026, 7, 23, 3, 0, 0, 0, time.UTC)
	bundle := buildRecoveryBundleFixture(t, now)
	store := &recoveryStoreFixture{records: make(map[int64]ports.SigningBundleRecord)}
	coordinator := NewRecoveryCoordinator(store, recoveryClockFixture{now: now})
	result, err := coordinator.RecordRecoverySigning(t.Context(), api.RecoverySigningRequest{
		EnvironmentID: bundle.EnvironmentID(), BundleBytes: bundle.CanonicalBytes()})
	if err != nil || result.Generation != 1 || result.SHA256 != bundle.SHA256() {
		t.Fatalf("record signing bundle = %#v, err=%v", result, err)
	}
	if _, err = coordinator.RecordRecoverySigning(t.Context(), api.RecoverySigningRequest{
		EnvironmentID: "018f0000-0000-7000-8000-000000000009",
		BundleBytes:   bundle.CanonicalBytes()}); err == nil {
		t.Fatal("cross-environment signing bundle was accepted")
	}
}
