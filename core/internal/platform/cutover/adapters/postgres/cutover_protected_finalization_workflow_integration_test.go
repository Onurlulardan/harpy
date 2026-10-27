//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func verifyProtectedFinalizationWorkflow(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	prepareIAMSystemResourcesAsyncSurfaces(t, ctx, database, "fresh")
	authorityBinary := buildCutoverAuthorityBinary(t, ctx)
	migratorBinary := buildCutoverMigratorBinary(t, ctx)
	directory := t.TempDir()
	now := time.Now().UTC()
	signer := prepareBuiltCutoverSigner(t, directory, now)
	seedReleaseReadiness(t, ctx, database.admin)
	addLastTenantCheckpoint(t, ctx, database.admin)
	assertObserverRequiresAdoptedCatalog(t, ctx, database)
	authority := buildBuiltAuthorityArtifacts(t, ctx, database, signer, directory, now)
	authority = adoptBuiltCutoverAuthority(
		t, ctx, authorityBinary, database, signer, authority, directory,
	)

	states := []struct {
		name, state string
		quiet       bool
	}{{"failed", "failed", false}, {"cancelled", "cancelled", false},
		{"completed", "completed", false}, {"completed_quiet", "completed", true}}
	for index, test := range states {
		t.Run(test.name, func(t *testing.T) {
			hooks := protectedFinalizationHooks{}
			if test.state == "completed" {
				hooks = completedIAMSystemResourcesHooks(t, ctx, database)
			}
			verifyProtectedFinalizationState(t, ctx, database, authorityBinary,
				migratorBinary, signer, authority, index+50, test.state, test.quiet,
				hooks)
		})
	}
}

func assertObserverRequiresAdoptedCatalog(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	setClusterAdminLease(t, ctx, database.admin, now.Add(110*time.Second))
	setFinalizerRoleLease(t, ctx, database.admin, now.Add(90*time.Second))
	connection := connectCutoverRole(t, ctx, database.dsn, "harpy_cluster_admin")
	_, err := connection.Exec(ctx, `SELECT * FROM
cutover_authority.observe_finalizer_authority_v1($1::uuid,$2::uuid)`,
		integrationOperationID, integrationRunID)
	_ = connection.Close(ctx)
	setCutoverRoleLease(t, ctx, database.admin, "cluster_admin_cleanup")
	cleanupBuiltFinalizerRole(t, ctx, database)
	if err == nil {
		t.Fatal("finalizer authority observation succeeded before catalog adoption")
	}
}

func verifyProtectedFinalizationState(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	authorityBinary, migratorBinary string, signer builtCutoverSigner,
	authority builtCutoverAuthorityArtifacts, identityIndex int, state string,
	quiet bool, hooks protectedFinalizationHooks,
) finalizationEvidence {
	t.Helper()
	directory := t.TempDir()
	now := time.Now().UTC()
	adoption := buildBuiltCutoverLeaseArtifacts(
		t, ctx, database, signer, directory,
		builtCutoverIdentityFor(identityIndex, state), now,
	)
	if quiet {
		adoption = useBuiltQuietCompletionBatch(t, signer, adoption, directory, now)
	}
	seedBuiltCutoverOperation(t, ctx, database, adoption)
	if quiet {
		removeBuiltJournalFixture(t, ctx, database, adoption)
	}
	leaseBuiltCutoverRoles(t, ctx, database, adoption.leaseExpiresAt(t))
	process := startBuiltCutoverProcess(
		t, ctx, migratorBinary, "cutover_adoption",
		builtAdoptionEnvironment(t, database, signer, adoption),
	)
	seal := publishBuiltSessionSealWithMutation(
		t, ctx, database, signer, adoption, process, nil,
	)
	outcome := waitBuiltCutoverProcess(t, process, 30*time.Second)
	assertBuiltCutoverResult(t, outcome, "cutover_adoption", 0)
	if state == "completed" {
		transitionBuiltOperationToQuiescing(t, ctx, database, adoption)
	}
	cleanup := publishBuiltCleanupReceipt(
		t, ctx, database, signer, adoption, seal, directory,
	)
	if state == "failed" {
		verifyProtectedCredentialReceiptRace(t, ctx, database, adoption, seal, cleanup)
	}
	finalizerDuration := 90 * time.Second
	if state == "failed" {
		finalizerDuration = 35 * time.Second
	}
	finalizer := issueBuiltFinalizerArtifacts(
		t, ctx, authorityBinary, database, signer, adoption,
		seal, cleanup, authority, directory, finalizerDuration,
	)
	evidence := finalizationEvidenceFromBuiltArtifacts(
		adoption, seal, cleanup, authority, finalizer,
	)
	evidence.migratorBinary = migratorBinary
	evidence.migratorEnvironment = builtFinalizationEnvironment(
		t, database, signer, adoption, seal, cleanup, finalizer, authority,
	)
	if hooks.beforeAuthorization != nil {
		hooks.beforeAuthorization(adoption, evidence)
	}
	assertMissingProtectedAuthorization(t, ctx, database, evidence)
	authorizeBuiltFinalization(t, ctx, authorityBinary, database, signer,
		adoption, seal, cleanup, authority, finalizer, directory)
	assertProtectedAuthorizationBindings(t, ctx, database, evidence)
	if state == "failed" {
		evidence = renewExpiredFinalizationAuthorization(t, ctx, database,
			authorityBinary, migratorBinary, signer, adoption, seal, cleanup, authority, evidence)
	}
	if hooks.afterAuthorization != nil {
		hooks.afterAuthorization(adoption, evidence)
	}
	verifyAuthorizedFinalization(t, ctx, database, evidence)
	if hooks.afterVerification != nil {
		hooks.afterVerification(adoption, evidence)
	}
	setCutoverRoleLease(t, ctx, database.admin, "cluster_admin_cleanup")
	cleanupBuiltFinalizerRole(t, ctx, database)
	return evidence
}

func assertProtectedAuthorizationBindings(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	evidence finalizationEvidence,
) {
	t.Helper()
	authorizer := connectCutoverRole(t, ctx, database.dsn, "harpy_cluster_admin")
	defer authorizer.Close(ctx)
	if err := authorizeFinalization(ctx, authorizer, evidence); err != nil {
		t.Fatalf("exact protected authorization replay failed: %v", err)
	}
	forged := evidence
	forged.sealDocument = replaceNestedJSONField(
		t, evidence.sealDocument, "payload", "receipt_id",
		"39000000-0000-4000-8000-000000000099",
	)
	if err := authorizeFinalization(ctx, authorizer, forged); err == nil {
		t.Fatal("forged protected authorization replay succeeded")
	}
}

func assertProtectedFinalizationTamper(
	t *testing.T, ctx context.Context, connection *pgx.Conn,
	evidence finalizationEvidence,
) {
	t.Helper()
	tampered := evidence
	tampered.terminalDocument = append(json.RawMessage(nil), evidence.terminalDocument...)
	tampered.terminalDocument[1] = ' '
	_, err := callFinalization(ctx, connection, tampered)
	assertFinalizationError(t, err, "22000", "authorization differs")
}

func assertProtectedReceiptVerification(
	t *testing.T, ctx context.Context, connection *pgx.Conn,
	evidence finalizationEvidence,
) {
	t.Helper()
	var verified bool
	err := connection.QueryRow(ctx, `SELECT
cutover_authority.verify_finalization_receipt_v1($1::uuid,$2::uuid,$3::uuid,$4::text,$5::text)`,
		evidence.operationID, evidence.runID, evidence.leaseID,
		evidence.terminalState, evidence.terminalSHA256,
	).Scan(&verified)
	if err != nil || !verified {
		t.Fatalf("verify protected finalization receipt: verified=%t err=%v", verified, err)
	}
}
