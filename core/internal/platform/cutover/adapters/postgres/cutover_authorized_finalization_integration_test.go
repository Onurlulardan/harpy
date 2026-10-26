//go:build integration

package postgres

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func verifyAuthorizedFinalization(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	evidence finalizationEvidence,
) {
	t.Helper()
	if evidence.terminalState == "completed" {
		verifyBuiltCompletedFinalization(t, ctx, database, evidence)
		return
	}
	verifyDirectAbortFinalization(t, ctx, database, evidence)
}

func verifyDirectAbortFinalization(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	evidence finalizationEvidence,
) {
	verifyCutoverRolePrivilegeBoundary(t, ctx, database, "harpy_cutover_finalizer")
	first := connectCutoverRole(t, ctx, database.dsn, "harpy_cutover_finalizer")
	second := connectFinalizerSession(t, ctx, database.dsn)
	verifyFinalizerLeaseDenials(t, ctx, database, first, evidence)
	outcomes := raceFinalization(ctx, first, second, evidence)
	assertExpectedFinalizationOutcomes(t, outcomes, evidence)
	assertFinalizationEvidence(t, ctx, database.admin, evidence)
	assertSingleProtectedFinalizationReceipt(t, ctx, database.admin, evidence)
	assertProtectedFinalizationTamper(t, ctx, first, evidence)
	assertProtectedReceiptVerification(t, ctx, first, evidence)
	_ = first.Close(ctx)
	_ = second.Close(ctx)
}

func verifyBuiltCompletedFinalization(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	evidence finalizationEvidence,
) {
	verifyCutoverRolePrivilegeBoundary(t, ctx, database, "harpy_cutover_finalizer")
	preflight := connectCutoverRole(t, ctx, database.dsn, "harpy_cutover_finalizer")
	verifyFinalizerLeaseDenials(t, ctx, database, preflight, evidence)
	_ = preflight.Close(ctx)
	leaseBuiltFinalizerRole(t, ctx, database, evidence.expiresAt)
	assertBuiltLeaseCatalog(t, ctx, database.admin, "harpy_cutover_finalizer",
		evidence.expiresAt, false, 1)
	assertBuiltRoleConnectionAccepted(t, ctx, database, "harpy_cutover_finalizer")
	// The first valid result is treated as lost after its DB commit, before publication.
	unpublished := runBuiltFinalization(t, ctx, database, evidence)
	replayed := runBuiltFinalization(t, ctx, database, evidence)
	if !bytes.Equal(unpublished.stdout, replayed.stdout) {
		t.Fatal("replayed cutover finalization stdout bytes differ")
	}
	assertFinalizationEvidence(t, ctx, database.admin, evidence)
	assertSingleProtectedFinalizationReceipt(t, ctx, database.admin, evidence)
	verifier, err := pgx.Connect(ctx, builtCutoverRoleDSN(
		t, database.dsn, "harpy_cutover_finalizer", builtCutoverPassword,
	))
	if err != nil {
		t.Fatalf("connect protected receipt verifier: %v", err)
	}
	assertProtectedFinalizationTamper(t, ctx, verifier, evidence)
	assertProtectedReceiptVerification(t, ctx, verifier, evidence)
	if err := verifier.Close(ctx); err != nil {
		t.Fatalf("close protected receipt verifier before built replay: %v", err)
	}
	drift := driftNormalOwnerFinalizationRows(t, ctx, database, evidence)
	defer restoreNormalOwnerFinalizationRows(t, ctx, database, drift)
	runBuiltFinalization(t, ctx, database, evidence)
	replayVerifier, err := pgx.Connect(ctx, builtCutoverRoleDSN(
		t, database.dsn, "harpy_cutover_finalizer", builtCutoverPassword,
	))
	if err != nil {
		t.Fatalf("reconnect protected receipt verifier: %v", err)
	}
	defer replayVerifier.Close(ctx)
	assertProtectedReceiptVerification(t, ctx, replayVerifier, evidence)
}

func runBuiltFinalization(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	evidence finalizationEvidence,
) builtCutoverOutcome {
	t.Helper()
	process := startBuiltCutoverProcess(t, ctx, evidence.migratorBinary,
		"cutover_finalization", evidence.migratorEnvironment)
	outcome := waitBuiltCutoverProcess(t, process, 30*time.Second)
	if outcome.exitCode != 0 {
		connection, err := pgx.Connect(ctx, builtCutoverRoleDSN(
			t, database.dsn, "harpy_cutover_finalizer", builtCutoverPassword,
		))
		if err == nil {
			_, diagnostic := callFinalization(ctx, connection, evidence)
			_ = connection.Close(ctx)
			t.Fatalf("built cutover_finalization exit=%d diagnostic=%v stderr=%s",
				outcome.exitCode, diagnostic, outcome.stderr)
		}
	}
	result := assertBuiltCutoverResult(t, outcome, "cutover_finalization", 0)
	if result.TerminalState == nil || *result.TerminalState != evidence.terminalState ||
		result.TerminalEvidenceSHA256 == nil ||
		*result.TerminalEvidenceSHA256 != evidence.terminalSHA256 {
		t.Fatalf("built finalization result differs: %+v", result)
	}
	return outcome
}

func assertExpectedFinalizationOutcomes(
	t *testing.T, outcomes [2]finalizationOutcome, evidence finalizationEvidence,
) {
	t.Helper()
	for _, outcome := range outcomes {
		if outcome.err != nil || outcome.state != evidence.terminalState ||
			outcome.evidenceSHA256 != evidence.terminalSHA256 {
			t.Fatalf("idempotent protected finalization = %#v", outcomes)
		}
	}
}
