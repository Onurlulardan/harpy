//go:build integration

package main

import (
	"strings"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

type workflowCutoverIDs struct {
	lease, operation, run, scope, seal, cleanup, finalizer string
}

func workflowIDsForState(state string) workflowCutoverIDs {
	if state == "completed" {
		return workflowCutoverIDs{
			lease:     "11111111-1111-4111-8111-111111111113",
			operation: "44444444-4444-4444-8444-444444444446",
			run:       "55555555-5555-4555-8555-555555555557",
			scope:     "66666666-6666-4666-8666-666666666668",
			seal:      "77777777-7777-4777-8777-777777777773",
			cleanup:   "88888888-8888-4888-8888-888888888883",
			finalizer: "99999999-9999-4999-8999-999999999993",
		}
	}
	if state == "cancelled" {
		return workflowCutoverIDs{
			lease:     "11111111-1111-4111-8111-111111111112",
			operation: "44444444-4444-4444-8444-444444444445",
			run:       "55555555-5555-4555-8555-555555555556",
			scope:     "66666666-6666-4666-8666-666666666667",
			seal:      "77777777-7777-4777-8777-777777777772",
			cleanup:   "88888888-8888-4888-8888-888888888882",
			finalizer: "99999999-9999-4999-8999-999999999992",
		}
	}
	return workflowCutoverIDs{
		lease:     "11111111-1111-4111-8111-111111111111",
		operation: "44444444-4444-4444-8444-444444444444",
		run:       "55555555-5555-4555-8555-555555555555",
		scope:     "66666666-6666-4666-8666-666666666666",
		seal:      "77777777-7777-4777-8777-777777777771",
		cleanup:   "88888888-8888-4888-8888-888888888881",
		finalizer: "99999999-9999-4999-8999-999999999991",
	}
}

func workflowCredentialLease(
	t *testing.T, target cutoverv1.DatabaseIdentity, ids workflowCutoverIDs, now time.Time,
) cutoverv1.CredentialLease {
	t.Helper()
	plan, transform := strings.Repeat("d", 64), strings.Repeat("e", 64)
	lease := cutoverv1.CredentialLease{
		ContractVersion: "cutover-credential-lease.v1", LeaseID: ids.lease,
		ReleaseID: workflowReleaseID, EnvironmentID: workflowEnvironmentID,
		OperationID: ids.operation, RunID: ids.run,
		IssuedAt:           now.Add(-2 * time.Minute).Format(time.RFC3339),
		NotBefore:          now.Add(-2 * time.Minute).Format(time.RFC3339),
		ExpiresAt:          now.Add(10 * time.Minute).Format(time.RFC3339),
		MaxDurationSeconds: 900, ScopeKind: "tenant", ScopeID: ids.scope,
		GroupID: "iam", Domains: []string{"iam"}, PlanSHA256: plan,
		TransformManifestSHA256: transform, Source: target, Target: target,
		PublicationName: "cutover_iam", SlotName: "cutover_iam_slot",
		SourceConnectionLimit: 2, TargetConnectionLimit: 1,
		RoleLeases: []cutoverv1.RoleLease{
			{Principal: "harpy_cutover_decoder", Login: true, Replication: true,
				ConnectionLimit: 2, PasswordState: "leased_secret"},
			{Principal: "harpy_adopter", Login: true,
				ConnectionLimit: 1, PasswordState: "leased_secret"},
		},
		Relations: []cutoverv1.RelationMapping{{
			Ordinal: 1, Source: cutoverv1.RelationIdentity{SchemaOID: 2200,
				RelationOID: 17000, SchemaName: "tenant_a", RelationName: "iam_source"},
			Target: cutoverv1.RelationIdentity{SchemaOID: 2200,
				RelationOID: 18000, SchemaName: "tenant_a", RelationName: "iam_target"},
			ReplicaIdentity: "primary_key", Columns: []cutoverv1.ColumnMapping{{
				Ordinal: 1, SourceAttnum: 1, TargetAttnum: 1,
				TransformID: "identity.uuid", PrimaryKey: true}},
			ProcedureContractID: "platform.cutover_probe.v1", ApplyProcedureOID: 19000,
			ApplyProcedureSignature: "tenant_a.apply_iam_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)",
			ExpectedCheckpointKeys:  []string{"iam/1"},
		}},
	}
	if err := cutoverv1.ValidateCredentialLease(lease); err != nil {
		t.Fatalf("validate workflow credential lease: %v", err)
	}
	return lease
}

func workflowTerminalEvidence(
	state string, ids workflowCutoverIDs, leaseSHA, planSHA, transformSHA string,
) cutoverv1.TerminalEvidence {
	if state == "completed" {
		finalLSN, snapshot := "0/20", "snapshot_1"
		checkpointCount, journalCount, appliedCount := 1, int64(1), int64(1)
		coverage := strings.Repeat("f", 64)
		return cutoverv1.TerminalEvidence{
			ContractVersion: "cutover-terminal-evidence.v1", OperationID: ids.operation,
			RunID: ids.run, LeaseManifestSHA256: leaseSHA, PlanSHA256: planSHA,
			TransformManifestSHA256: transformSHA, TerminalState: state,
			FinalLSN: &finalLSN, InitialCopy: &cutoverv1.TerminalInitialCopyEvidence{
				ManifestSHA256: leaseSHA, SnapshotName: snapshot, SnapshotLSN: "0/10",
				RelationCount: 1, PartitionCount: 1, RowCount: 1,
				RootSHA256: strings.Repeat("a", 64)},
			CheckpointCount: &checkpointCount,
			Checkpoints: []cutoverv1.TerminalCheckpointEvidence{{
				RelationOrdinal: 1, CheckpointKey: "iam/1", AppliedLSN: "0/20",
				AppliedRows: 1, BatchCount: 1, BatchRootSHA256: strings.Repeat("b", 64)}},
			JournalTransactionCount: &journalCount, AppliedChangeCount: &appliedCount,
			JournalCoverageRootSHA256: &coverage,
		}
	}
	code := "verification_failed"
	if state == "cancelled" {
		code = "operator_cancelled"
	}
	outcome := strings.Repeat("f", 64)
	return cutoverv1.TerminalEvidence{
		ContractVersion: "cutover-terminal-evidence.v1", OperationID: ids.operation,
		RunID: ids.run, LeaseManifestSHA256: leaseSHA, PlanSHA256: planSHA,
		TransformManifestSHA256: transformSHA, TerminalState: state,
		OutcomeCode: &code, OutcomeEvidenceSHA256: &outcome,
	}
}
