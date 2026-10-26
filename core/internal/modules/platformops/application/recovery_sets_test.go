package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestRequestRecoverySetDerivesCurrentOnPremiseIdentity(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	bundle, _, _ := signingFixture(t, now)
	store := &recoveryControlFixture{recoveryStoreFixture: recoveryStoreFixture{
		records: map[int64]ports.SigningBundleRecord{1: {
			EnvironmentID: appUUID(1), Generation: 1, SHA256: bundle.SHA256(),
			BundleBytes: bundle.CanonicalBytes(), CreatedAt: bundle.GeneratedAt()}}},
		sets: map[string]ports.RecoverySetRecord{}, restores: map[string]ports.RestoreRecord{},
		identity: appIdentity()}
	service := NewRecoveryService(store, store, store, store, recoveryClockFixture{now: now})
	request := api.RecoverySetRequest{RecoverySetID: appUUID(6), OperationID: appUUID(5),
		EnvironmentID: appUUID(1)}
	created, err := service.RequestRecoverySet(context.Background(), request)
	if err != nil || created.Sequence != 1 || created.State != "requested" {
		t.Fatalf("request result = %#v %v", created, err)
	}
	request.EnvironmentID = appUUID(9)
	if _, err = service.RequestRecoverySet(context.Background(), request); err == nil {
		t.Fatal("caller-selected environment was accepted")
	}
}

func TestPruneRecoverySetsKeepsExactPolicyAndLegalHold(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	store := &recoveryControlFixture{sets: map[string]ports.RecoverySetRecord{},
		restores: map[string]ports.RestoreRecord{}}
	for index := 0; index < 16; index++ {
		store.retention = append(store.retention, ports.RetentionRecord{
			RecoverySetID: fmt.Sprintf("set-%02d", index), ManifestSHA256: appDigest("a"),
			ProtectedAt: now.Add(-time.Duration(index) * 8 * 24 * time.Hour)})
	}
	store.retention = append(store.retention, ports.RetentionRecord{
		RecoverySetID: "legal-hold", ManifestSHA256: appDigest("b"),
		ProtectedAt: now.AddDate(-5, 0, 0), LegalHold: true})
	service := NewRecoveryService(store, store, store, store, recoveryClockFixture{now: now})
	decision, err := service.PruneRecoverySets(context.Background(), appUUID(1))
	if err != nil || len(decision.Keep) != 11 || len(decision.Delete) != 6 {
		t.Fatalf("retention = %d/%d %v", len(decision.Keep), len(decision.Delete), err)
	}
}
