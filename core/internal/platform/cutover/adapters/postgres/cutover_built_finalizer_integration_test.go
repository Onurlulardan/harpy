//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

const builtFinalizerProcedure = "cutover_authority.finalize_operation_v1(" +
	"uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)"

type builtCutoverAuthorityArtifacts struct {
	releaseArtifact    builtCutoverArtifact
	releaseSignature   builtCutoverArtifact
	bootstrapAuthority builtCutoverArtifact
	bootstrapSignature builtCutoverArtifact
	operationArtifact  builtCutoverArtifact
	operationSignature builtCutoverArtifact
	receiptArtifact    builtCutoverArtifact
	receiptSignature   builtCutoverArtifact
}

type builtCutoverFinalizerArtifacts struct {
	leaseArtifact    builtCutoverArtifact
	leaseSignature   builtCutoverArtifact
	terminalArtifact builtCutoverArtifact
	expiresAt        time.Time
}

func builtAdoptionEnvironment(
	t *testing.T, database cutoverIntegrationDatabase, signer builtCutoverSigner,
	artifacts builtCutoverLeaseArtifacts,
) []string {
	t.Helper()
	return []string{
		"HARPY_ENVIRONMENT=test", "HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_CUTOVER_SOURCE_DSN=" + builtCutoverRoleDSN(
			t, database.dsn, "harpy_cutover_decoder", builtCutoverPassword),
		"HARPY_CUTOVER_TARGET_DSN=" + builtCutoverRoleDSN(
			t, database.dsn, "harpy_adopter", builtCutoverPassword),
		"HARPY_POSTGRES_CONNECT_TIMEOUT=5s", "HARPY_POSTGRES_CLEANUP_TIMEOUT=5s",
		"HARPY_LOG_LEVEL=error", "HARPY_CUTOVER_LEASE_MANIFEST_PATH=" +
			artifacts.leaseArtifact.path,
		"HARPY_CUTOVER_LEASE_SIGNATURE_PATH=" + artifacts.leaseSignature.path,
		"HARPY_CUTOVER_TRUST_ROOT_PATH=" + signer.rootPath,
		"HARPY_CUTOVER_SESSION_SEAL_RECEIPT_PATH=" + artifacts.sealPath,
		"HARPY_CUTOVER_BATCH_PATH=" + artifacts.batchArtifact.path,
	}
}

func builtFinalizationEnvironment(
	t *testing.T, database cutoverIntegrationDatabase, signer builtCutoverSigner,
	adoption builtCutoverLeaseArtifacts, seal, cleanup builtCutoverArtifact,
	finalizer builtCutoverFinalizerArtifacts, authority builtCutoverAuthorityArtifacts,
) []string {
	t.Helper()
	return []string{
		"HARPY_ENVIRONMENT=test", "HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_CUTOVER_FINALIZER_DSN=" + builtCutoverRoleDSN(
			t, database.dsn, "harpy_cutover_finalizer", builtCutoverPassword),
		"HARPY_POSTGRES_CONNECT_TIMEOUT=5s", "HARPY_POSTGRES_CLEANUP_TIMEOUT=5s",
		"HARPY_LOG_LEVEL=error",
		"HARPY_CUTOVER_LEASE_MANIFEST_PATH=" + adoption.leaseArtifact.path,
		"HARPY_CUTOVER_LEASE_SIGNATURE_PATH=" + adoption.leaseSignature.path,
		"HARPY_CUTOVER_TRUST_ROOT_PATH=" + signer.rootPath,
		"HARPY_CUTOVER_SESSION_SEAL_RECEIPT_PATH=" + seal.path,
		"HARPY_CUTOVER_CREDENTIAL_CLEANUP_RECEIPT_PATH=" + cleanup.path,
		"HARPY_CUTOVER_FINALIZER_LEASE_MANIFEST_PATH=" + finalizer.leaseArtifact.path,
		"HARPY_CUTOVER_FINALIZER_LEASE_SIGNATURE_PATH=" + finalizer.leaseSignature.path,
		"HARPY_RELEASE_MANIFEST_PATH=" + authority.releaseArtifact.path,
		"HARPY_RELEASE_SIGNATURE_PATH=" + authority.releaseSignature.path,
		"HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH=" + authority.bootstrapAuthority.path,
		"HARPY_BOOTSTRAP_AUTHORITY_SIGNATURE_PATH=" + authority.bootstrapSignature.path,
		"HARPY_CUTOVER_AUTHORITY_OPERATION_PATH=" + authority.operationArtifact.path,
		"HARPY_CUTOVER_AUTHORITY_OPERATION_SIGNATURE_PATH=" + authority.operationSignature.path,
		"HARPY_CUTOVER_AUTHORITY_RECEIPT_PATH=" + authority.receiptArtifact.path,
		"HARPY_CUTOVER_AUTHORITY_RECEIPT_SIGNATURE_PATH=" + authority.receiptSignature.path,
		"HARPY_CUTOVER_TERMINAL_EVIDENCE_PATH=" + finalizer.terminalArtifact.path,
	}
}

func leaseAndAssertBuiltFinalizer(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	artifacts builtCutoverFinalizerArtifacts,
) {
	t.Helper()
	leaseBuiltFinalizerRole(t, ctx, database, artifacts.expiresAt)
	assertBuiltLeaseCatalog(
		t, ctx, database.admin, "harpy_cutover_finalizer",
		artifacts.expiresAt, false, 1,
	)
	assertBuiltRoleConnectionAccepted(t, ctx, database, "harpy_cutover_finalizer")
}
