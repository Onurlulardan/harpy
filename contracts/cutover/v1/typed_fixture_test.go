package cutoverv1

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

const (
	typedLeaseID       = "11111111-1111-4111-8111-111111111111"
	typedReleaseID     = "22222222-2222-4222-8222-222222222222"
	typedEnvironmentID = "33333333-3333-4333-8333-333333333333"
	typedOperationID   = "44444444-4444-4444-8444-444444444444"
	typedRunID         = "55555555-5555-4555-8555-555555555555"
)

func validTypedCredentialLease() CredentialLease {
	digest := strings.Repeat("a", 64)
	database := validTypedDatabaseIdentity()
	return CredentialLease{
		ContractVersion: "cutover-credential-lease.v1",
		LeaseID:         typedLeaseID, ReleaseID: typedReleaseID, EnvironmentID: typedEnvironmentID,
		OperationID: typedOperationID, RunID: typedRunID,
		IssuedAt: "2026-07-13T12:00:00Z", NotBefore: "2026-07-13T12:00:00Z",
		ExpiresAt: "2026-07-13T12:15:00Z", MaxDurationSeconds: 900,
		ScopeKind: "tenant", ScopeID: "66666666-6666-4666-8666-666666666666",
		GroupID: "iam_system_resources", Domains: []string{"iam", "system_resource_layout"},
		PlanSHA256:              digest,
		TransformManifestSHA256: digest, Source: database, Target: database,
		PublicationName: "cutover_iam", SlotName: "cutover_iam_slot",
		SourceConnectionLimit: 2, TargetConnectionLimit: 1,
		RoleLeases: []RoleLease{
			{Principal: "harpy_cutover_decoder", Login: true, Replication: true, ConnectionLimit: 2, PasswordState: "leased_secret"},
			{Principal: "harpy_adopter", Login: true, ConnectionLimit: 1, PasswordState: "leased_secret"},
		},
		Relations: []RelationMapping{validTypedRelation()},
	}
}

func validTypedDatabaseIdentity() DatabaseIdentity {
	return DatabaseIdentity{
		DatabaseName: "harpy", DatabaseOrigin: "v1_adopted",
		SystemIdentifier: "123456789", DatabaseOID: 16384,
	}
}

func validTypedRelation() RelationMapping {
	return RelationMapping{
		Ordinal:         1,
		Source:          RelationIdentity{SchemaOID: 2200, RelationOID: 17000, SchemaName: "tenant_a", RelationName: "iam_source"},
		Target:          RelationIdentity{SchemaOID: 2200, RelationOID: 18000, SchemaName: "tenant_a", RelationName: "iam_target"},
		ReplicaIdentity: "primary_key",
		Columns: []ColumnMapping{{
			Ordinal: 1, SourceAttnum: 1, TargetAttnum: 1,
			TransformID: "identity.uuid", PrimaryKey: true,
		}},
		ProcedureContractID: "platform.cutover_probe.v1", ApplyProcedureOID: 19000,
		ApplyProcedureSignature: "tenant_a.apply_iam_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)",
		ExpectedCheckpointKeys:  []string{"iam/1"},
	}
}

func validTypedSignature() Signature {
	return Signature{
		Algorithm: "Ed25519", KeyID: "cutover-2026",
		Value: base64.RawURLEncoding.EncodeToString(make([]byte, 64)),
	}
}

func validTypedSessionSeal(leaseSHA string) SessionSealReceipt {
	applicationName := "harpy-cutover/" + typedRunID
	database := validTypedDatabaseIdentity()
	return SessionSealReceipt{
		ContractVersion: "cutover-session-seal-receipt.v1", Kind: "cutover_session_seal",
		Payload: SessionSealPayload{
			ReceiptID: "77777777-7777-4777-8777-777777777777",
			LeaseID:   typedLeaseID, LeaseManifestSHA256: leaseSHA,
			EnvironmentID: typedEnvironmentID, OperationID: typedOperationID, RunID: typedRunID,
			IssuedAt: "2026-07-13T12:02:00Z", Source: database, Target: database,
			Credentials: []SealedCredential{
				{Principal: "harpy_cutover_decoder"}, {Principal: "harpy_adopter"},
			},
			SourceSessions: []CutoverSession{
				{PID: 101, BackendStart: "2026-07-13T12:01:00Z", SessionUser: "harpy_cutover_decoder", ApplicationName: applicationName},
				{PID: 102, BackendStart: "2026-07-13T12:01:01Z", SessionUser: "harpy_cutover_decoder", ApplicationName: applicationName},
			},
			TargetSessions: []CutoverSession{
				{PID: 201, BackendStart: "2026-07-13T12:01:02Z", SessionUser: "harpy_adopter", ApplicationName: applicationName},
			},
			Result: "sealed",
		},
		Signature: validTypedSignature(),
	}
}

func validTypedCleanup(leaseSHA, sealSHA string) CredentialCleanupReceipt {
	return CredentialCleanupReceipt{
		ContractVersion: "cutover-credential-cleanup-receipt.v1",
		Kind:            "cutover_credential_cleanup",
		Payload: CredentialCleanupPayload{
			ReceiptID: "88888888-8888-4888-8888-888888888888",
			LeaseID:   typedLeaseID, LeaseManifestSHA256: leaseSHA, SealReceiptSHA256: sealSHA,
			EnvironmentID: typedEnvironmentID, OperationID: typedOperationID, RunID: typedRunID,
			IssuedAt: "2026-07-13T12:03:00Z",
			Principals: []CleanedPrincipal{
				{Principal: "harpy_cutover_decoder", CleanedAt: "2026-07-13T12:02:59Z", DynamicGrantsRevoked: true},
				{Principal: "harpy_adopter", CleanedAt: "2026-07-13T12:03:00Z", DynamicGrantsRevoked: true},
			},
			Result: "succeeded",
		},
		Signature: validTypedSignature(),
	}
}

func validTypedChangeBatch() ChangeBatch {
	return ChangeBatch{
		ContractVersion: "cutover-change-batch.v1",
		Payload: ChangeBatchPayload{
			OperationID: typedOperationID, RunID: typedRunID,
			ScopeKind: "tenant", ScopeID: "66666666-6666-4666-8666-666666666666",
			RelationOrdinal: 1, CheckpointKey: "iam/1", CheckpointState: "delta",
			Position: SourcePosition{
				SnapshotName: "snapshot_1", SnapshotLSN: "0/10", ReceivedLSN: "0/20",
				CommitLSN: "0/20", SourceXID: "42", BatchOrdinal: 1,
			},
			LastKey: json.RawMessage(`{"id":"1"}`),
			Changes: []Change{{
				Ordinal: 1, Kind: "insert", Key: json.RawMessage(`{"id":"1"}`),
				Before: json.RawMessage("null"), After: json.RawMessage(`{"id":"1","name":"Ada"}`),
			}},
		},
		BatchSHA256: strings.Repeat("c", 64),
	}
}

func validTypedFinalizerLease(adoptionSHA, cleanupSHA string) FinalizerLease {
	tenantDigest := strings.Repeat("f", 64)
	tenantCount := 3
	return FinalizerLease{
		ContractVersion: "cutover-finalizer-lease.v1",
		LeaseID:         "99999999-9999-4999-8999-999999999999", ReleaseID: typedReleaseID,
		EnvironmentID: typedEnvironmentID, OperationID: typedOperationID, RunID: typedRunID,
		IssuedAt: "2026-07-13T12:03:00Z", NotBefore: "2026-07-13T12:03:00Z",
		ExpiresAt: "2026-07-13T12:05:00Z", MaxDurationSeconds: 120,
		RoleLease: RoleLease{
			Principal: "harpy_cutover_finalizer", Login: true,
			ConnectionLimit: 1, PasswordState: "leased_secret",
		},
		Target: validTypedDatabaseIdentity(), AdoptionLeaseSHA256: adoptionSHA,
		CleanupReceiptSHA256: cleanupSHA, AuthorityAdoptionReceiptSHA256: strings.Repeat("e", 64),
		FinalizationStateSHA256:        strings.Repeat("9", 64),
		TenantRequirementCatalogSHA256: &tenantDigest, TenantRequirementCount: &tenantCount,
		FinalizationProcedureOID:       20000,
		FinalizationProcedureSignature: finalizationProcedureSignature,
		TerminalState:                  "completed", TerminalEvidenceSHA256: strings.Repeat("d", 64),
	}
}

func validTypedTerminalEvidence() TerminalEvidence {
	digest := strings.Repeat("a", 64)
	finalLSN := "0/20"
	initial := TerminalInitialCopyEvidence{
		ManifestSHA256: digest, SnapshotName: "snapshot_1", SnapshotLSN: "0/10",
		RelationCount: 1, PartitionCount: 1, RowCount: 1, RootSHA256: digest,
	}
	checkpointCount, journalCount, appliedCount := 1, int64(1), int64(1)
	return TerminalEvidence{
		ContractVersion: "cutover-terminal-evidence.v1",
		OperationID:     typedOperationID, RunID: typedRunID,
		LeaseManifestSHA256: digest, PlanSHA256: digest,
		TransformManifestSHA256: digest, TerminalState: "completed", FinalLSN: &finalLSN,
		InitialCopy: &initial, CheckpointCount: &checkpointCount,
		Checkpoints: []TerminalCheckpointEvidence{{
			RelationOrdinal: 1, CheckpointKey: "iam/1", AppliedLSN: "0/20",
			AppliedRows: 1, BatchCount: 1, BatchRootSHA256: digest,
		}},
		JournalTransactionCount: &journalCount, AppliedChangeCount: &appliedCount,
		JournalCoverageRootSHA256: &digest,
	}
}
