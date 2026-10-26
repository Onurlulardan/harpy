//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type builtCutoverIdentity struct {
	operationID    string
	runID          string
	leaseID        string
	finalizerLease string
	sealReceiptID  string
	cleanupReceipt string
	scopeID        string
	terminalState  string
}

type builtCutoverLeaseArtifacts struct {
	identity       builtCutoverIdentity
	lease          cutoverv1.CredentialLease
	leaseArtifact  builtCutoverArtifact
	leaseSignature builtCutoverArtifact
	batchArtifact  builtCutoverArtifact
	sealPath       string
}

func builtCutoverIdentityFor(index int, state string) builtCutoverIdentity {
	prefix := fmt.Sprintf("%08d", 31000000+index)
	uuid := func(suffix int) string {
		return fmt.Sprintf("%s-0000-4000-8000-%012d", prefix, suffix)
	}
	return builtCutoverIdentity{
		operationID: uuid(1), runID: uuid(2), leaseID: uuid(3), finalizerLease: uuid(4),
		sealReceiptID: uuid(5), cleanupReceipt: uuid(6), scopeID: uuid(7),
		terminalState: state,
	}
}

func buildBuiltCutoverLeaseArtifacts(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	signer builtCutoverSigner, directory string, identity builtCutoverIdentity, now time.Time,
) builtCutoverLeaseArtifacts {
	t.Helper()
	now = now.UTC().Truncate(time.Microsecond)
	databaseIdentity, procedureOID := readBuiltCutoverDatabaseIdentity(t, ctx, database)
	relation := readBuiltCutoverRelation(t, ctx, database, procedureOID)
	issuedAt := now.Add(-time.Second)
	expiresAt := now.Add(8 * time.Minute)
	lease := cutoverv1.CredentialLease{
		ContractVersion: "cutover-credential-lease.v1",
		LeaseID:         identity.leaseID, ReleaseID: integrationReleaseID,
		EnvironmentID: integrationEnvironment, OperationID: identity.operationID,
		RunID: identity.runID, IssuedAt: issuedAt.Format(time.RFC3339Nano),
		NotBefore: now.Add(-500 * time.Millisecond).Format(time.RFC3339Nano),
		ExpiresAt: expiresAt.Format(time.RFC3339Nano), MaxDurationSeconds: 900,
		ScopeKind: "tenant", ScopeID: identity.scopeID, GroupID: "iam_system_resources",
		Domains: []string{"iam", "system_resource_layout"}, PlanSHA256: digestMarker('b'),
		TransformManifestSHA256: digestMarker('c'), Source: databaseIdentity,
		Target: databaseIdentity, PublicationName: "cutover_iam",
		SlotName: "cutover_iam_slot", SourceConnectionLimit: 2, TargetConnectionLimit: 1,
		RoleLeases: []cutoverv1.RoleLease{
			{Principal: "harpy_cutover_decoder", Login: true, Replication: true,
				ConnectionLimit: 2, PasswordState: "leased_secret"},
			{Principal: "harpy_adopter", Login: true,
				ConnectionLimit: 1, PasswordState: "leased_secret"},
		},
		Relations: []cutoverv1.RelationMapping{relation},
	}
	if err := cutoverv1.ValidateCredentialLease(lease); err != nil {
		t.Fatalf("validate built cutover lease: %v", err)
	}
	manifest := writeBuiltCutoverArtifact(t, directory, "lease.json", lease)
	signature := cutoverv1.CredentialLeaseSignature{
		ContractVersion: "cutover-credential-lease-signature.v1",
		ManifestSHA256:  manifest.digest,
		Signature:       signBuiltCutoverDocument(t, signer, manifest.document, issuedAt),
	}
	signed := writeBuiltCutoverArtifact(t, directory, "lease.sig", signature)
	batch := buildBuiltCutoverBatch(t, identity)
	batchArtifact := writeBuiltCutoverArtifact(t, directory, "batch.json", batch)
	assertBuiltAdoptionArtifactsVerify(t, signer, manifest, signed, batchArtifact, now)
	return builtCutoverLeaseArtifacts{
		identity: identity, lease: lease, leaseArtifact: manifest,
		leaseSignature: signed, batchArtifact: batchArtifact,
		sealPath: filepath.Join(directory, "session-seal.json"),
	}
}

func assertBuiltAdoptionArtifactsVerify(
	t *testing.T, signer builtCutoverSigner,
	lease, signature, batch builtCutoverArtifact, now time.Time,
) {
	t.Helper()
	verifiedLease, err := attestation.VerifyCutoverCredentialLease(
		lease.document, signature.document, signer.root, now,
	)
	if err != nil {
		t.Fatalf("verify built cutover lease fixture: %v", err)
	}
	if _, err := attestation.VerifyCutoverChangeBatch(batch.document, verifiedLease); err != nil {
		t.Fatalf("verify built cutover batch fixture: %v", err)
	}
}

func readBuiltCutoverRelation(
	t *testing.T,
	ctx context.Context,
	database cutoverIntegrationDatabase,
	procedureOID uint32,
) cutoverv1.RelationMapping {
	t.Helper()
	var sourceSchema, sourceRelation, targetSchema, targetRelation uint32
	err := database.admin.QueryRow(ctx, `SELECT
    source_namespace.oid, source_relation.oid,
    target_namespace.oid, target_relation.oid
FROM pg_catalog.pg_namespace AS source_namespace
JOIN pg_catalog.pg_class AS source_relation
  ON source_relation.relnamespace=source_namespace.oid
CROSS JOIN pg_catalog.pg_namespace AS target_namespace
JOIN pg_catalog.pg_class AS target_relation
  ON target_relation.relnamespace=target_namespace.oid
WHERE source_namespace.nspname='legacy_source'
  AND source_relation.relname='allowed_rows'
  AND target_namespace.nspname='platform'
  AND target_relation.relname='integration_business_rows'`).Scan(
		&sourceSchema, &sourceRelation, &targetSchema, &targetRelation,
	)
	if err != nil {
		t.Fatalf("read built cutover relation identity: %v", err)
	}
	return cutoverv1.RelationMapping{
		Ordinal: 1,
		Source: cutoverv1.RelationIdentity{SchemaOID: sourceSchema, RelationOID: sourceRelation,
			SchemaName: "legacy_source", RelationName: "allowed_rows"},
		Target: cutoverv1.RelationIdentity{SchemaOID: targetSchema, RelationOID: targetRelation,
			SchemaName: "platform", RelationName: "integration_business_rows"},
		ReplicaIdentity: "primary_key",
		Columns: []cutoverv1.ColumnMapping{{Ordinal: 1, SourceAttnum: 1,
			TargetAttnum: 1, TransformID: "identity.uuid", PrimaryKey: true}},
		ProcedureContractID: "platform.cutover_probe.v1", ApplyProcedureOID: procedureOID,
		ApplyProcedureSignature: "platform.cutover_probe_apply_v1(uuid,uuid,text,uuid," +
			"smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)",
		ExpectedCheckpointKeys: []string{"main"},
	}
}

func buildBuiltCutoverBatch(
	t *testing.T, identity builtCutoverIdentity,
) cutoverv1.ChangeBatch {
	t.Helper()
	finalLSN := "0/20"
	payload := cutoverv1.ChangeBatchPayload{
		OperationID: identity.operationID, RunID: identity.runID,
		ScopeKind: "tenant", ScopeID: identity.scopeID, RelationOrdinal: 1,
		CheckpointKey: "main", CheckpointState: "complete",
		Position: cutoverv1.SourcePosition{SnapshotName: "snapshot-main",
			SnapshotLSN: "0/10", ReceivedLSN: "0/20", CommitLSN: "0/20",
			FinalLSN: &finalLSN, SourceXID: "41", BatchOrdinal: 1},
		LastKey: json.RawMessage(`{"id":1}`),
		Changes: []cutoverv1.Change{{Ordinal: 1, Kind: "insert",
			Key: json.RawMessage(`{"id":1}`), Before: json.RawMessage("null"),
			After: json.RawMessage(`{"id":1}`)}},
	}
	document, err := attestation.MarshalCanonical(payload)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := attestation.DigestHex(document)
	if err != nil {
		t.Fatal(err)
	}
	batch := cutoverv1.ChangeBatch{
		ContractVersion: "cutover-change-batch.v1", Payload: payload, BatchSHA256: digest,
	}
	if err := cutoverv1.ValidateChangeBatch(batch); err != nil {
		t.Fatalf("validate built cutover batch: %v", err)
	}
	return batch
}
