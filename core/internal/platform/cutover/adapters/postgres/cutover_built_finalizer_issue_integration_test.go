//go:build integration

package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func issueBuiltFinalizerArtifacts(
	t *testing.T, ctx context.Context, binary string, database cutoverIntegrationDatabase,
	signer builtCutoverSigner, adoption builtCutoverLeaseArtifacts,
	seal, cleanup builtCutoverArtifact, authority builtCutoverAuthorityArtifacts,
	directory string, requestedDuration ...time.Duration,
) builtCutoverFinalizerArtifacts {
	t.Helper()
	terminal := buildBuiltTerminalEvidence(t, ctx, database, adoption, directory)
	now := time.Now().UTC().Truncate(time.Microsecond)
	duration := 90 * time.Second
	if len(requestedDuration) == 1 {
		duration = requestedDuration[0]
	}
	expiresAt := now.Add(duration)
	setClusterAdminLease(t, ctx, database.admin, now.Add(110*time.Second))
	setFinalizerRoleLease(t, ctx, database.admin, expiresAt)
	assertBuiltObserverAvailable(t, ctx, database, adoption)
	clusterDSN := builtCutoverRoleDSN(
		t, database.dsn, "harpy_cluster_admin", "integration-leased-secret",
	)
	finalizerDSN := builtCutoverRoleDSN(
		t, database.dsn, "harpy_cutover_finalizer", "integration-leased-secret",
	)
	clusterPath := writeBuiltSecret(t, directory, "issuer-cluster-admin.dsn", clusterDSN)
	finalizerPath := writeBuiltSecret(t, directory, "issuer-finalizer.dsn", finalizerDSN)
	leasePath := filepath.Join(directory, "finalizer-lease.json")
	signaturePath := filepath.Join(directory, "finalizer-lease.sig")
	profile := builtFinalizationPrerequisites(
		signer, adoption, seal, cleanup, authority, terminal,
	)
	profile["cluster_admin_dsn_file"] = clusterPath
	profile["finalizer_dsn_file"] = finalizerPath
	profile["signing_key_file"] = signer.signingKeyPath
	profile["finalizer_lease_id"] = adoption.identity.finalizerLease
	profile["finalizer_lease_path"] = leasePath
	profile["finalizer_lease_signature_path"] = signaturePath
	config := writeBuiltCutoverArtifact(t, directory, "issue-finalizer-config.json", map[string]any{
		"contract_version": "external-cutover-authority-config.v1",
		"action":           "issue_finalizer", "environment": "test", "deadline_seconds": 120,
		"issue_finalizer": profile,
	})
	firstOutput, err := runBuiltAuthorityProcess(ctx, binary, config.path)
	if err != nil {
		diagnostic := diagnoseBuiltFinalizerObserver(
			t, ctx, database, adoption, seal, cleanup, expiresAt,
		)
		t.Fatalf("run external cutover authority: %v: %s observer=%s",
			err, firstOutput, diagnostic)
	}
	assertBuiltFinalizerIssueReplay(t, firstOutput, false)
	firstLease := readBuiltArtifact(t, leasePath)
	firstSignature := readBuiltArtifact(t, signaturePath)
	if err := os.Remove(signaturePath); err != nil {
		t.Fatalf("remove issued signature at publication crash seam: %v", err)
	}
	assertBuiltFinalizerIssueReplay(
		t, runBuiltAuthorityCommand(t, ctx, binary, config.path), false,
	)
	recoveredLease := readBuiltArtifact(t, leasePath)
	recoveredSignature := readBuiltArtifact(t, signaturePath)
	if !bytes.Equal(firstLease.document, recoveredLease.document) ||
		!bytes.Equal(firstSignature.document, recoveredSignature.document) {
		t.Fatal("finalizer publication recovery changed signed artifact bytes")
	}
	assertBuiltFinalizerIssueReplay(
		t, runBuiltAuthorityCommand(t, ctx, binary, config.path), true,
	)
	return builtCutoverFinalizerArtifacts{
		leaseArtifact: recoveredLease, leaseSignature: recoveredSignature,
		terminalArtifact: terminal, expiresAt: expiresAt,
	}
}

func assertBuiltObserverAvailable(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	adoption builtCutoverLeaseArtifacts,
) {
	t.Helper()
	connection := connectCutoverRole(t, ctx, database.dsn, "harpy_cluster_admin")
	var stateSHA256, procedureSignature, databaseName, databaseOrigin string
	var systemIdentifier, operationState string
	var procedureOID, databaseOID uint32
	var completedEligible, abortEligible bool
	var tenantCatalogSHA256 *string
	var tenantRequirementCount *int
	var finalizerValidUntil time.Time
	row := connection.QueryRow(ctx, `SELECT * FROM
cutover_authority.observe_finalizer_authority_v1($1::uuid,$2::uuid)`,
		adoption.identity.operationID, adoption.identity.runID)
	err := row.Scan(
		&stateSHA256, &procedureOID, &procedureSignature, &databaseName, &databaseOrigin,
		&systemIdentifier, &databaseOID, &operationState, &completedEligible,
		&abortEligible, &tenantCatalogSHA256, &tenantRequirementCount, &finalizerValidUntil,
	)
	_ = connection.Close(ctx)
	if err != nil {
		var postgresError *pgconn.PgError
		_ = errors.As(err, &postgresError)
		t.Fatalf("protected observer diagnostic: %v detail=%q where=%q",
			err, postgresError.Detail, postgresError.Where)
	}
}

func assertBuiltFinalizerIssueReplay(t *testing.T, document []byte, expected bool) {
	t.Helper()
	var result struct {
		Action   string `json:"action"`
		Replayed bool   `json:"replayed"`
	}
	if err := json.Unmarshal(document, &result); err != nil ||
		result.Action != "issue_finalizer" || result.Replayed != expected {
		t.Fatalf("finalizer issuance replay result = %s err=%v", document, err)
	}
}

func authorizeBuiltFinalization(
	t *testing.T, ctx context.Context, binary string, database cutoverIntegrationDatabase,
	signer builtCutoverSigner, adoption builtCutoverLeaseArtifacts,
	seal, cleanup builtCutoverArtifact, authority builtCutoverAuthorityArtifacts,
	finalizer builtCutoverFinalizerArtifacts, directory string,
) {
	t.Helper()
	clusterDSN := builtCutoverRoleDSN(
		t, database.dsn, "harpy_cluster_admin", "integration-leased-secret",
	)
	clusterPath := writeBuiltSecret(t, directory, "authorizer-cluster-admin.dsn", clusterDSN)
	profile := builtFinalizationPrerequisites(
		signer, adoption, seal, cleanup, authority, finalizer.terminalArtifact,
	)
	profile["cluster_admin_dsn_file"] = clusterPath
	profile["finalizer_lease"] = finalizer.leaseArtifact.path
	profile["finalizer_lease_signature"] = finalizer.leaseSignature.path
	config := writeBuiltCutoverArtifact(t, directory, "authorize-finalization-config.json", map[string]any{
		"contract_version": "external-cutover-authority-config.v1",
		"action":           "authorize_finalization", "environment": "test", "deadline_seconds": 120,
		"authorize_finalization": profile,
	})
	runBuiltAuthorityCommand(t, ctx, binary, config.path)
	runBuiltAuthorityCommand(t, ctx, binary, config.path)
}

func builtFinalizationPrerequisites(
	signer builtCutoverSigner, adoption builtCutoverLeaseArtifacts,
	seal, cleanup builtCutoverArtifact, authority builtCutoverAuthorityArtifacts,
	terminal builtCutoverArtifact,
) map[string]any {
	return map[string]any{
		"trust_root":                    signer.rootPath,
		"release_manifest":              authority.releaseArtifact.path,
		"release_signature":             authority.releaseSignature.path,
		"bootstrap_authority_manifest":  authority.bootstrapAuthority.path,
		"bootstrap_authority_signature": authority.bootstrapSignature.path,
		"authority_operation":           authority.operationArtifact.path,
		"authority_operation_signature": authority.operationSignature.path,
		"authority_receipt":             authority.receiptArtifact.path,
		"authority_receipt_signature":   authority.receiptSignature.path,
		"credential_lease":              adoption.leaseArtifact.path,
		"credential_lease_signature":    adoption.leaseSignature.path,
		"session_seal_receipt":          seal.path,
		"credential_cleanup_receipt":    cleanup.path,
		"terminal_evidence":             terminal.path,
	}
}
