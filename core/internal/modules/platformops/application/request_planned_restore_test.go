package application

import (
	"context"
	"crypto/ed25519"
	"testing"
	"time"

	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type restorePlanSignerFixture struct {
	bundle  recoverysigningv1.Bundle
	private ed25519.PrivateKey
}

func newRestorePlanSignerFixture(t *testing.T, bundleBytes, seed []byte) *restorePlanSignerFixture {
	t.Helper()
	bundle, err := recoverysigningv1.Parse(bundleBytes)
	if err != nil {
		t.Fatal(err)
	}
	return &restorePlanSignerFixture{bundle: bundle, private: ed25519.NewKeyFromSeed(seed)}
}

func (signer *restorePlanSignerFixture) SignRestorePlan(
	ctx context.Context, request ports.RestorePlanSigningRequest,
) (ports.RestorePlanSigningResult, error) {
	artifact, err := environmentrecoveryv1.Build(environmentrecoveryv1.Input{
		Purpose:       environmentrecoveryv1.PurposeRestorePlan,
		EnvironmentID: request.EnvironmentID, OperationID: request.OperationID,
		ArtifactID: request.OperationID, Sequence: request.Sequence,
		IssuedAt: request.IssuedAt, ExpiresAt: request.ExpiresAt, Payload: request.Payload,
	}, signer.bundle, signer.private)
	if err != nil {
		return ports.RestorePlanSigningResult{}, err
	}
	return ports.RestorePlanSigningResult{Generation: signer.bundle.Generation(),
		Document: artifact.CanonicalBytes()}, nil
}

func TestRequestPlannedRestoreBuildsSignedSessionBoundPlan(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	store, signedRequest, controlKey := restoreServiceFixture(t, now)
	signer := newRestorePlanSignerFixture(t, store.records[1].BundleBytes, controlKey.Seed())
	service := NewRecoveryServiceWithSigner(store, store, store, store,
		signer, recoveryClockFixture{now: now})
	request := plannedRequestFromSigned(signedRequest)
	operation, err := service.RequestPlannedRestore(context.Background(), request)
	if err != nil || operation.State != "plan_ready" ||
		operation.RecoverySetID != request.RecoverySetID {
		t.Fatalf("planned restore = %#v / %v", operation, err)
	}
	stored, found := store.restores[request.OperationID]
	if !found || stored.PlanSHA256 == "" {
		t.Fatal("signed restore plan was not durably requested")
	}
}

func TestRequestPlannedRestoreReplaysPersistedOperation(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	store, signedRequest, controlKey := restoreServiceFixture(t, now)
	base := NewRecoveryService(store, store, store, store, recoveryClockFixture{now: now})
	if _, err := base.RequestRestore(context.Background(), signedRequest); err != nil {
		t.Fatal(err)
	}
	signer := newRestorePlanSignerFixture(t, store.records[1].BundleBytes, controlKey.Seed())
	service := NewRecoveryServiceWithSigner(store, store, store, store,
		signer, recoveryClockFixture{now: now.Add(time.Minute)})
	operation, err := service.RequestPlannedRestore(
		context.Background(), plannedRequestFromSigned(signedRequest))
	if err != nil || operation.State != "plan_ready" {
		t.Fatalf("planned restore replay = %#v / %v", operation, err)
	}
}

func plannedRequestFromSigned(request api.RestoreOperationRequest) api.PlannedRestoreRequest {
	return api.PlannedRestoreRequest{OperationID: request.OperationID,
		RecoverySetID:        request.RecoverySetID,
		RequestedByTenantID:  request.RequestedByTenantID,
		RequestedByActorID:   request.RequestedByActorID,
		RequestedBySessionID: request.RequestedBySessionID,
		Disposition:          api.TargetExistingReadable,
		SafetyRecoverySetID:  appUUID(7), SafetyRecoverySetSHA256: appDigest("7")}
}
