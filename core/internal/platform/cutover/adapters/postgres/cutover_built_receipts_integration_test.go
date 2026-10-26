//go:build integration

package postgres

import (
	"context"
	"sort"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func publishBuiltSessionSeal(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	signer builtCutoverSigner, artifacts builtCutoverLeaseArtifacts,
) builtCutoverArtifact {
	return publishBuiltSessionSealWithMutation(
		t, ctx, database, signer, artifacts, nil, nil,
	)
}

func publishBuiltSessionSealWithMutation(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	signer builtCutoverSigner, artifacts builtCutoverLeaseArtifacts,
	process *builtCutoverProcess,
	mutate func(*cutoverv1.SessionSealReceipt),
) builtCutoverArtifact {
	t.Helper()
	application := cutoverAdoptionApplicationName(artifacts.identity.runID)
	source, target := awaitBuiltCutoverSessions(t, ctx, database, application, process)
	sealBuiltCutoverRoles(t, ctx, database)
	issuedAt := time.Now().UTC()
	receipt := cutoverv1.SessionSealReceipt{
		ContractVersion: "cutover-session-seal-receipt.v1", Kind: "cutover_session_seal",
		Payload: cutoverv1.SessionSealPayload{
			ReceiptID: artifacts.identity.sealReceiptID, LeaseID: artifacts.identity.leaseID,
			LeaseManifestSHA256: artifacts.leaseArtifact.digest,
			EnvironmentID:       integrationEnvironment, OperationID: artifacts.identity.operationID,
			RunID: artifacts.identity.runID, IssuedAt: issuedAt.Format(time.RFC3339Nano),
			Source: artifacts.lease.Source, Target: artifacts.lease.Target,
			Credentials: []cutoverv1.SealedCredential{
				{Principal: "harpy_cutover_decoder", ConnectionLimit: 0},
				{Principal: "harpy_adopter", ConnectionLimit: 0},
			},
			SourceSessions: source, TargetSessions: target, Result: "sealed",
		},
	}
	if mutate != nil {
		mutate(&receipt)
	}
	receipt.Signature = signBuiltCutoverEnvelope(t, signer, receipt.SigningEnvelope(), issuedAt)
	if err := cutoverv1.ValidateSessionSealReceipt(receipt); err != nil {
		t.Fatalf("validate built session seal: %v", err)
	}
	return publishBuiltCutoverArtifact(t, artifacts.sealPath, receipt)
}

func awaitBuiltCutoverSessions(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	application string, process *builtCutoverProcess,
) ([]cutoverv1.CutoverSession, []cutoverv1.CutoverSession) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if process != nil {
			select {
			case err := <-process.done:
				t.Fatalf("built cutover exited before seal: %v stdout=%s stderr=%s",
					err, process.stdout.Bytes(), process.stderr.Bytes())
			default:
			}
		}
		source, target, err := readBuiltCutoverSessions(ctx, database, application)
		if err == nil && len(source) == 2 && len(target) == 1 {
			return source, target
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("built cutover process did not open its bounded session set")
	return nil, nil
}

func readBuiltCutoverSessions(
	ctx context.Context, database cutoverIntegrationDatabase, application string,
) ([]cutoverv1.CutoverSession, []cutoverv1.CutoverSession, error) {
	rows, err := database.admin.Query(ctx, `SELECT pid,backend_start,usename,application_name
FROM pg_catalog.pg_stat_activity
WHERE application_name=$1
  AND usename IN ('harpy_cutover_decoder','harpy_adopter')`, application)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var source, target []cutoverv1.CutoverSession
	for rows.Next() {
		var session cutoverv1.CutoverSession
		var backendStart time.Time
		if err := rows.Scan(&session.PID, &backendStart, &session.SessionUser,
			&session.ApplicationName); err != nil {
			return nil, nil, err
		}
		session.BackendStart = backendStart.UTC().Format(time.RFC3339Nano)
		if session.SessionUser == "harpy_cutover_decoder" {
			source = append(source, session)
		} else {
			target = append(target, session)
		}
	}
	sort.Slice(source, func(left, right int) bool { return source[left].PID < source[right].PID })
	return source, target, rows.Err()
}

func publishBuiltCleanupReceipt(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	signer builtCutoverSigner, artifacts builtCutoverLeaseArtifacts,
	seal builtCutoverArtifact, directory string,
) builtCutoverArtifact {
	t.Helper()
	cleanupBuiltAdoptionRoles(t, ctx, database)
	issuedAt := time.Now().UTC()
	principals := []cutoverv1.CleanedPrincipal{
		{Principal: "harpy_cutover_decoder", CleanedAt: issuedAt.Format(time.RFC3339Nano),
			ConnectionLimit: 0, DynamicGrantsRevoked: true},
		{Principal: "harpy_adopter", CleanedAt: issuedAt.Format(time.RFC3339Nano),
			ConnectionLimit: 0, DynamicGrantsRevoked: true},
	}
	receipt := cutoverv1.CredentialCleanupReceipt{
		ContractVersion: "cutover-credential-cleanup-receipt.v1",
		Kind:            "cutover_credential_cleanup",
		Payload: cutoverv1.CredentialCleanupPayload{
			ReceiptID:           artifacts.identity.cleanupReceipt,
			LeaseID:             artifacts.identity.leaseID,
			LeaseManifestSHA256: artifacts.leaseArtifact.digest,
			SealReceiptSHA256:   seal.digest, EnvironmentID: integrationEnvironment,
			OperationID: artifacts.identity.operationID, RunID: artifacts.identity.runID,
			IssuedAt: issuedAt.Format(time.RFC3339Nano), Principals: principals,
			Result: "succeeded",
		},
	}
	receipt.Signature = signBuiltCutoverEnvelope(t, signer, receipt.SigningEnvelope(), issuedAt)
	if err := cutoverv1.ValidateCredentialCleanupReceipt(receipt); err != nil {
		t.Fatalf("validate built cleanup receipt: %v", err)
	}
	return writeBuiltCutoverArtifact(t, directory, "credential-cleanup.json", receipt)
}
