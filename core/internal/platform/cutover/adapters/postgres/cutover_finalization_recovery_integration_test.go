//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func renewExpiredFinalizationAuthorization(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	authorityBinary, migratorBinary string, signer builtCutoverSigner,
	adoption builtCutoverLeaseArtifacts, seal, cleanup builtCutoverArtifact,
	authority builtCutoverAuthorityArtifacts, previous finalizationEvidence,
) finalizationEvidence {
	t.Helper()
	candidateAdoption := adoption
	candidateAdoption.identity.finalizerLease = adoption.identity.operationID[:8] +
		"-0000-4000-8000-000000000044"
	candidateFinalizer := issueBuiltFinalizerArtifacts(t, ctx, authorityBinary,
		database, signer, candidateAdoption, seal, cleanup, authority, t.TempDir())
	candidate := finalizationEvidenceFromBuiltArtifacts(
		candidateAdoption, seal, cleanup, authority, candidateFinalizer,
	)
	authorizer := connectCutoverRole(t, ctx, database.dsn, "harpy_cluster_admin")
	err := authorizeFinalization(ctx, authorizer, candidate)
	assertFinalizationError(t, err, "55000", "authorization attempt is active")
	_ = authorizer.Close(ctx)
	waitForFinalizationAuthorizationExpiry(t, ctx, previous.expiresAt)
	renewedAdoption := adoption
	renewedAdoption.identity.finalizerLease = adoption.identity.operationID[:8] +
		"-0000-4000-8000-000000000045"
	renewedFinalizer := issueBuiltFinalizerArtifacts(t, ctx, authorityBinary,
		database, signer, renewedAdoption, seal, cleanup, authority, t.TempDir())
	renewed := finalizationEvidenceFromBuiltArtifacts(
		renewedAdoption, seal, cleanup, authority, renewedFinalizer,
	)
	renewed.migratorBinary = migratorBinary
	renewed.migratorEnvironment = builtFinalizationEnvironment(t, database, signer,
		renewedAdoption, seal, cleanup, renewedFinalizer, authority)
	authorizer = connectCutoverRole(t, ctx, database.dsn, "harpy_cluster_admin")
	defer authorizer.Close(ctx)
	if err := authorizeFinalization(ctx, authorizer, renewed); err != nil {
		t.Fatalf("authorize renewed finalization attempt: %v", err)
	}
	if err := authorizeFinalization(ctx, authorizer, renewed); err != nil {
		t.Fatalf("replay renewed finalization attempt: %v", err)
	}
	assertFinalizationAuthorizationHistory(t, ctx, database.admin, previous, renewed)
	return renewed
}

func waitForFinalizationAuthorizationExpiry(
	t *testing.T, ctx context.Context, expiresAt time.Time,
) {
	t.Helper()
	delay := time.Until(expiresAt) + 500*time.Millisecond
	if delay <= 0 {
		return
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		t.Fatalf("wait for real finalization authorization expiry: %v", ctx.Err())
	case <-timer.C:
	}
}

func assertFinalizationAuthorizationHistory(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
	previous, renewed finalizationEvidence,
) {
	t.Helper()
	var attempts int
	var exactLeases bool
	err := admin.QueryRow(ctx, `SELECT count(*),bool_and(finalizer_lease_id IN ($3,$4))
FROM cutover_authority.finalization_authorizations
WHERE operation_id=$1 AND run_id=$2`, previous.operationID, previous.runID,
		previous.leaseID, renewed.leaseID).Scan(&attempts, &exactLeases)
	if err != nil || attempts != 2 || !exactLeases {
		t.Fatalf("protected authorization history: attempts=%d exact=%t err=%v",
			attempts, exactLeases, err)
	}
}
