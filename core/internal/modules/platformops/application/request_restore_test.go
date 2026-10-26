package application

import (
	"context"
	"crypto/ed25519"
	"testing"
	"time"

	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestRequestRestoreAcceptsOnlyPersistedSignedSameEnvironmentPlan(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	fixture, request, _ := restoreServiceFixture(t, now)
	service := NewRecoveryService(fixture, fixture, fixture, fixture,
		recoveryClockFixture{now: now})
	operation, err := service.RequestRestore(context.Background(), request)
	if err != nil || operation.State != "plan_ready" {
		t.Fatalf("restore request failed: %#v %v", operation, err)
	}
	tampered := request
	tampered.OperationID = appUUID(9)
	tampered.PlanBytes = append([]byte(nil), request.PlanBytes...)
	tampered.PlanBytes[len(tampered.PlanBytes)-2] ^= 1
	if _, err = service.RequestRestore(context.Background(), tampered); err == nil {
		t.Fatal("tampered restore plan accepted")
	}
}

func restoreServiceFixture(
	t *testing.T, now time.Time,
) (*recoveryControlFixture, api.RestoreOperationRequest, ed25519.PrivateKey) {
	t.Helper()
	bundle, manifestKey, controlKey := signingFixture(t, now)
	identity := appIdentity()
	manifest := appManifest(identity, now)
	manifest.RecoveryTrustBundleSHA256 = bundle.SHA256()
	manifestPayload, _ := manifest.CanonicalBytes()
	manifestArtifact, err := environmentrecoveryv1.Build(environmentrecoveryv1.Input{
		Purpose:       environmentrecoveryv1.PurposeRecoveryManifest,
		EnvironmentID: identity.EnvironmentID, OperationID: appUUID(5),
		ArtifactID: appUUID(6), Sequence: 1, IssuedAt: now,
		ExpiresAt: now.Add(24 * time.Hour), Payload: manifestPayload}, bundle, manifestKey)
	if err != nil {
		t.Fatal(err)
	}
	safetyID := appUUID(7)
	gate := recoveryset.RestoreRequest{RecoverySetID: appUUID(6),
		Manifest: manifest, Target: identity, Disposition: recoveryset.TargetExistingReadable,
		Safety: recoveryset.SafetyEvidence{RecoverySetID: safetyID,
			RecoverySetSHA256: appDigest("7")}}
	planPayload, err := recoveryset.BuildRestorePlan(gate, appUUID(6), manifestArtifact.SHA256())
	if err != nil {
		t.Fatal(err)
	}
	planArtifact, err := environmentrecoveryv1.Build(environmentrecoveryv1.Input{
		Purpose:       environmentrecoveryv1.PurposeRestorePlan,
		EnvironmentID: identity.EnvironmentID, OperationID: appUUID(8),
		ArtifactID: appUUID(8), Sequence: 1, IssuedAt: now,
		ExpiresAt: now.Add(time.Hour), Payload: planPayload}, bundle, controlKey)
	if err != nil {
		t.Fatal(err)
	}
	store := &recoveryControlFixture{recoveryStoreFixture: recoveryStoreFixture{
		records: map[int64]ports.SigningBundleRecord{1: {
			EnvironmentID: identity.EnvironmentID, Generation: 1, SHA256: bundle.SHA256(),
			BundleBytes: bundle.CanonicalBytes(), CreatedAt: bundle.GeneratedAt()}}},
		sets: map[string]ports.RecoverySetRecord{}, restores: map[string]ports.RestoreRecord{},
		identity: identity}
	store.sets[appUUID(6)] = ports.RecoverySetRecord{RecoverySetID: appUUID(6),
		OperationID: appUUID(5), EnvironmentID: identity.EnvironmentID,
		Sequence: 1, State: "protected", ManifestSHA256: manifestArtifact.SHA256(),
		ManifestBytes: manifestArtifact.CanonicalBytes(), DatabaseSystemID: manifest.DatabaseSystemID,
		RecoverySigningGeneration: 1, RecoveryTrustBundleSHA256: bundle.SHA256()}
	store.sets[safetyID] = ports.RecoverySetRecord{RecoverySetID: safetyID,
		EnvironmentID: identity.EnvironmentID, State: "protected",
		ManifestSHA256: appDigest("7")}
	return store, api.RestoreOperationRequest{OperationID: appUUID(8),
		RecoverySetID: appUUID(6), RequestedByTenantID: appUUID(4),
		RequestedByActorID:   appUUID(2),
		RequestedBySessionID: appUUID(3), PlanBytes: planArtifact.CanonicalBytes(),
		PlanSigningGeneration: 1}, controlKey
}
