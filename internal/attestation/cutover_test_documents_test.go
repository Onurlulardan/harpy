package attestation

import (
	"strings"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

const (
	attestationLeaseID       = "11111111-1111-4111-8111-111111111111"
	attestationReleaseID     = "22222222-2222-4222-8222-222222222222"
	attestationEnvironmentID = "33333333-3333-4333-8333-333333333333"
	attestationOperationID   = "44444444-4444-4444-8444-444444444444"
	attestationRunID         = "55555555-5555-4555-8555-555555555555"
)

func attestationCutoverLease() cutoverv1.CredentialLease {
	digest := strings.Repeat("a", 64)
	database := attestationCutoverDatabase()
	return cutoverv1.CredentialLease{
		ContractVersion: "cutover-credential-lease.v1",
		LeaseID:         attestationLeaseID, ReleaseID: attestationReleaseID,
		EnvironmentID: attestationEnvironmentID, OperationID: attestationOperationID,
		RunID: attestationRunID, IssuedAt: "2026-07-13T12:00:00Z",
		NotBefore: "2026-07-13T12:00:00Z", ExpiresAt: "2026-07-13T12:15:00Z",
		MaxDurationSeconds: 900, ScopeKind: "tenant",
		ScopeID: "66666666-6666-4666-8666-666666666666",
		GroupID: "iam_system_resources",
		Domains: []string{"iam", "system_resource_layout"}, PlanSHA256: digest,
		TransformManifestSHA256: digest, Source: database, Target: database,
		PublicationName: "cutover_iam", SlotName: "cutover_iam_slot",
		SourceConnectionLimit: 2, TargetConnectionLimit: 1,
		RoleLeases: []cutoverv1.RoleLease{
			{Principal: "harpy_cutover_decoder", Login: true, Replication: true, ConnectionLimit: 2, PasswordState: "leased_secret"},
			{Principal: "harpy_adopter", Login: true, ConnectionLimit: 1, PasswordState: "leased_secret"},
		},
		Relations: []cutoverv1.RelationMapping{{
			Ordinal:         1,
			Source:          cutoverv1.RelationIdentity{SchemaOID: 2200, RelationOID: 17000, SchemaName: "tenant_a", RelationName: "iam_source"},
			Target:          cutoverv1.RelationIdentity{SchemaOID: 2200, RelationOID: 18000, SchemaName: "tenant_a", RelationName: "iam_target"},
			ReplicaIdentity: "primary_key",
			Columns: []cutoverv1.ColumnMapping{{
				Ordinal: 1, SourceAttnum: 1, TargetAttnum: 1,
				TransformID: "identity.uuid", PrimaryKey: true,
			}},
			ProcedureContractID: "platform.cutover_probe.v1", ApplyProcedureOID: 19000,
			ApplyProcedureSignature: "tenant_a.apply_iam_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)",
			ExpectedCheckpointKeys:  []string{"main"},
		}},
	}
}

func attestationCutoverDatabase() cutoverv1.DatabaseIdentity {
	return cutoverv1.DatabaseIdentity{
		DatabaseName: "harpy", DatabaseOrigin: "v1_adopted",
		SystemIdentifier: "123456789", DatabaseOID: 16384,
	}
}

func attestationSessionSeal(leaseSHA string) cutoverv1.SessionSealReceipt {
	application := "harpy-cutover/" + attestationRunID
	database := attestationCutoverDatabase()
	return cutoverv1.SessionSealReceipt{
		ContractVersion: "cutover-session-seal-receipt.v1", Kind: "cutover_session_seal",
		Payload: cutoverv1.SessionSealPayload{
			ReceiptID: "77777777-7777-4777-8777-777777777777",
			LeaseID:   attestationLeaseID, LeaseManifestSHA256: leaseSHA,
			EnvironmentID: attestationEnvironmentID, OperationID: attestationOperationID,
			RunID: attestationRunID, IssuedAt: "2026-07-13T12:02:00Z",
			Source: database, Target: database,
			Credentials: []cutoverv1.SealedCredential{
				{Principal: "harpy_cutover_decoder"}, {Principal: "harpy_adopter"},
			},
			SourceSessions: []cutoverv1.CutoverSession{
				{PID: 101, BackendStart: "2026-07-13T12:01:00Z", SessionUser: "harpy_cutover_decoder", ApplicationName: application},
				{PID: 102, BackendStart: "2026-07-13T12:01:01Z", SessionUser: "harpy_cutover_decoder", ApplicationName: application},
			},
			TargetSessions: []cutoverv1.CutoverSession{
				{PID: 201, BackendStart: "2026-07-13T12:01:02Z", SessionUser: "harpy_adopter", ApplicationName: application},
			},
			Result: "sealed",
		},
	}
}

func attestationCleanup(leaseSHA, sealSHA string) cutoverv1.CredentialCleanupReceipt {
	return cutoverv1.CredentialCleanupReceipt{
		ContractVersion: "cutover-credential-cleanup-receipt.v1",
		Kind:            "cutover_credential_cleanup",
		Payload: cutoverv1.CredentialCleanupPayload{
			ReceiptID: "88888888-8888-4888-8888-888888888888",
			LeaseID:   attestationLeaseID, LeaseManifestSHA256: leaseSHA,
			SealReceiptSHA256: sealSHA, EnvironmentID: attestationEnvironmentID,
			OperationID: attestationOperationID, RunID: attestationRunID,
			IssuedAt: "2026-07-13T12:03:00Z",
			Principals: []cutoverv1.CleanedPrincipal{
				{Principal: "harpy_cutover_decoder", CleanedAt: "2026-07-13T12:02:59Z", DynamicGrantsRevoked: true},
				{Principal: "harpy_adopter", CleanedAt: "2026-07-13T12:03:00Z", DynamicGrantsRevoked: true},
			},
			Result: "succeeded",
		},
	}
}
