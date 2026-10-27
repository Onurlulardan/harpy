//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type finalizationEvidence struct {
	operationID, runID, leaseID         string
	leaseSHA256                         string
	adoptionSHA256, sealSHA256          string
	cleanupSHA256, authoritySHA256      string
	terminalState, terminalSHA256       string
	expiresAt                           time.Time
	leaseDocument, adoptionDocument     json.RawMessage
	sealDocument, cleanupDocument       json.RawMessage
	authorityDocument, terminalDocument json.RawMessage
	migratorBinary                      string
	migratorEnvironment                 []string
}

func finalizationEvidenceFromBuiltArtifacts(
	adoption builtCutoverLeaseArtifacts, seal, cleanup builtCutoverArtifact,
	authority builtCutoverAuthorityArtifacts, finalizer builtCutoverFinalizerArtifacts,
) finalizationEvidence {
	clone := func(document []byte) json.RawMessage {
		return append(json.RawMessage(nil), document...)
	}
	return finalizationEvidence{
		operationID: adoption.identity.operationID, runID: adoption.identity.runID,
		leaseID:        adoption.identity.finalizerLease,
		leaseSHA256:    finalizer.leaseArtifact.digest,
		adoptionSHA256: adoption.leaseArtifact.digest,
		sealSHA256:     seal.digest, cleanupSHA256: cleanup.digest,
		authoritySHA256:  authority.receiptArtifact.digest,
		terminalState:    adoption.identity.terminalState,
		terminalSHA256:   finalizer.terminalArtifact.digest,
		expiresAt:        finalizer.expiresAt,
		leaseDocument:    clone(finalizer.leaseArtifact.document),
		adoptionDocument: clone(adoption.leaseArtifact.document),
		sealDocument:     clone(seal.document), cleanupDocument: clone(cleanup.document),
		authorityDocument: clone(authority.receiptArtifact.document),
		terminalDocument:  clone(finalizer.terminalArtifact.document),
	}
}

func mustIntegrationJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func seedReleaseReadiness(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	_, err := admin.Exec(ctx, `
INSERT INTO platform.release_generations (
    migration_run_id, release_id, environment_id, database_origin,
    database_generation, minimum_runtime_generation,
    migration_manifest_sha256, extension_manifest_sha256
) VALUES ($1,$2,$3,'fresh',2,1,$4,$5);
INSERT INTO platform.schema_contracts (
    component, verified_run_id, database_generation, minimum_runtime_generation,
    migration_manifest_sha256, extension_manifest_sha256, status
) VALUES ('core',$1,2,1,$4,$5,'ready');
INSERT INTO platform.tenants (id, schema_name, deleted_at)
VALUES ($6,'tenant_alpha',NULL),($7,'tenant_beta',NULL);
INSERT INTO platform.tenant_migration_checkpoints
    (tenant_id, owner, logical_version, database_origin, selected_path,
     selected_sha256, run_id, state)
VALUES
($6,'tenancy',20260714011000,'fresh','core/internal/modules/tenancy/migrations/tenant/fresh/20260714011000_create_tenancy_schema_identity.fresh.sql','3639feb1de870e0d71490f5530391fb952db381dadabd94574c53c7f113bf471',$8,'succeeded'),
($6,'tenancy',20260714011100,'fresh','core/internal/modules/tenancy/migrations/tenant/fresh/20260714011100_create_tenancy_schema_identity_guard.fresh.sql','193a74b5dd839f9e6e2e9f8c24bc16fe0e2e190ab1fc0e78b034417ad47d4250',$8,'succeeded'),
($6,'tenancy',20260714011200,'fresh','core/internal/modules/tenancy/migrations/tenant/fresh/20260714011200_guard_tenancy_schema_identity.fresh.sql','7c40e9e01ebdf6e9033f27f26eb04f82a93775db29f46d9f8760c6b9c352e479',$8,'succeeded'),
($7,'tenancy',20260714011000,'fresh','core/internal/modules/tenancy/migrations/tenant/fresh/20260714011000_create_tenancy_schema_identity.fresh.sql','3639feb1de870e0d71490f5530391fb952db381dadabd94574c53c7f113bf471',$8,'succeeded'),
($7,'tenancy',20260714011100,'fresh','core/internal/modules/tenancy/migrations/tenant/fresh/20260714011100_create_tenancy_schema_identity_guard.fresh.sql','193a74b5dd839f9e6e2e9f8c24bc16fe0e2e190ab1fc0e78b034417ad47d4250',$8,'succeeded');`,
		pgx.QueryExecModeSimpleProtocol,
		integrationReleaseRun, integrationReleaseID, integrationEnvironment,
		digestMarker('6'), digestMarker('7'), integrationTenantID,
		integrationTenantBID, integrationPriorRun,
	)
	if err != nil {
		t.Fatalf("seed release readiness: %v", err)
	}
	seedTenantRequirementAuthority(t, ctx, admin)
}

func addLastTenantCheckpoint(t *testing.T, ctx context.Context, admin *pgx.Conn) {
	t.Helper()
	_, err := admin.Exec(ctx, `INSERT INTO platform.tenant_migration_checkpoints
    (tenant_id, owner, logical_version, database_origin, selected_path,
     selected_sha256, run_id, state)
VALUES ($1,'tenancy',20260714011200,'fresh',
    'core/internal/modules/tenancy/migrations/tenant/fresh/20260714011200_guard_tenancy_schema_identity.fresh.sql',
    '7c40e9e01ebdf6e9033f27f26eb04f82a93775db29f46d9f8760c6b9c352e479',$2,'succeeded')`,
		integrationTenantBID, integrationReleaseRun)
	if err != nil {
		t.Fatal(err)
	}
}

func setTenantCheckpointPath(
	t *testing.T, ctx context.Context, admin *pgx.Conn, path string,
) {
	t.Helper()
	_, err := admin.Exec(ctx, `UPDATE platform.tenant_migration_checkpoints
SET selected_path=$1 WHERE tenant_id=$2 AND owner='tenancy' AND logical_version=20260714011000`,
		path, integrationTenantID)
	if err != nil {
		t.Fatal(err)
	}
}

func setTenantCheckpointHash(
	t *testing.T, ctx context.Context, admin *pgx.Conn, digest string,
) {
	t.Helper()
	_, err := admin.Exec(ctx, `UPDATE platform.tenant_migration_checkpoints
SET selected_sha256=$1 WHERE tenant_id=$2 AND owner='tenancy' AND logical_version=20260714011000`,
		digest, integrationTenantID)
	if err != nil {
		t.Fatal(err)
	}
}

func connectFinalizerSession(
	t *testing.T,
	ctx context.Context,
	dsn string,
) *pgx.Conn {
	t.Helper()
	connection, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := connection.Exec(ctx, "SET SESSION AUTHORIZATION harpy_cutover_finalizer"); err != nil {
		_ = connection.Close(ctx)
		t.Fatal(err)
	}
	return connection
}
