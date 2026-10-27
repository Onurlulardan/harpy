//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/fullrollout"
	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/grouprehearsal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func seedFullRolloutAuthority(
	t *testing.T, ctx context.Context, admin *pgx.Conn, environmentID, tenantID uuid.UUID,
) {
	t.Helper()
	requester := uuid.MustParse(integrationOwnerID)
	approver := uuid.MustParse(integrationRunID)
	domains := []string{"iam", "system_resource_layout", "billing", "workflow",
		"analytics", "catalog", "work", "documents", "notifications"}
	_, err := admin.Exec(ctx, `INSERT INTO platform.domain_cutovers (
scope_kind,scope_id,domain,state,writer,generation,requested_by,approved_by,
reason,change_reference,acceptance_digest,changed_at
) VALUES ('global',$1,'billing','v2','v2',1,$3,$4,
'full rollout PG18','PHASE13-FINAL',repeat('a',64),clock_timestamp()-interval '1 hour');
INSERT INTO platform.domain_cutovers (
scope_kind,scope_id,domain,state,writer,generation,requested_by,approved_by,
reason,change_reference,acceptance_digest,changed_at
) SELECT 'tenant',$2,domain,'v2','v2',1,$3,$4,
'full rollout PG18','PHASE13-FINAL',repeat('a',64),clock_timestamp()-interval '1 hour'
FROM pg_catalog.unnest($5::text[]) AS domain;
INSERT INTO platform.cutover_groups (
scope_kind,scope_id,group_id,state,writer,generation,requested_by,approved_by,
reason,change_reference,acceptance_digest,changed_at
) VALUES
('tenant',$2,'billing','v2','v2',1,$3,$4,
'full rollout PG18','PHASE13-FINAL',repeat('b',64),clock_timestamp()-interval '1 hour'),
('tenant',$2,'iam_system_resources','v2','v2',1,$3,$4,
'full rollout PG18','PHASE13-FINAL',repeat('c',64),clock_timestamp()-interval '1 hour');
INSERT INTO platform.cutover_group_members
(scope_kind,scope_id,group_id,ordinal,domain) VALUES
('tenant',$2,'billing',1,'billing'),
('tenant',$2,'iam_system_resources',1,'iam'),
('tenant',$2,'iam_system_resources',2,'system_resource_layout');
WITH moment AS (SELECT pg_catalog.clock_timestamp() AS value)
UPDATE platform.contract_retirements SET state='eligible',remaining_scope_count=0,
zero_hit_since=moment.value-interval '31 days',eligible_at=moment.value,
revision=revision+1,updated_at=moment.value FROM moment;
WITH moment AS (SELECT pg_catalog.clock_timestamp() AS value)
UPDATE platform.contract_retirements SET state='retired',retired_at=moment.value,
revision=revision+1,updated_at=moment.value FROM moment`,
		pgx.QueryExecModeSimpleProtocol,
		environmentID, tenantID, requester, approver, domains)
	if err != nil {
		t.Fatalf("seed full rollout authority: %v", err)
	}
}

func fullRolloutCompletionFixture(
	environmentID, tenantID, receiptID uuid.UUID, now time.Time,
) fullrollout.PersistedCompletion {
	digest := sha256.Sum256([]byte("phase13 verified artifact"))
	requester := uuid.MustParse(integrationOwnerID)
	approver := uuid.MustParse(integrationRunID)
	steps := make([]fullrollout.StepEvidence, 0, len(fullrollout.RequiredSteps()))
	for _, step := range fullrollout.RequiredSteps() {
		scopeID := tenantID
		if step.Scope == fullrollout.ScopeGlobal {
			scopeID = environmentID
		}
		steps = append(steps, fullrollout.StepEvidence{Step: step, ScopeID: scopeID,
			RequestedBy: requester, ApprovedBy: approver, Generation: 1,
			State: "v2", Writer: "v2", AcceptanceSHA256: digest,
			ReverseRehearsalSHA256: digest, SourceCount: 12, TargetCount: 12,
			SourceSHA256: digest, TargetSHA256: digest})
	}
	order := grouprehearsal.CohortOrder()
	var total time.Duration
	for _, cohort := range order {
		minimum, _ := grouprehearsal.MinimumSoak(cohort)
		total += minimum
	}
	cursor := now.Add(-total - time.Hour)
	cohorts := make([]grouprehearsal.SoakEvidence, 0, len(order))
	for _, cohort := range order {
		minimum, _ := grouprehearsal.MinimumSoak(cohort)
		end := cursor.Add(minimum)
		cohorts = append(cohorts, grouprehearsal.SoakEvidence{Cohort: cohort,
			StartedAt: cursor, EndedAt: end,
			EligibleRequests: 1_000, ComparedRequests: 1_000})
		cursor = end
	}
	evidence := fullrollout.CompletionEvidence{Steps: steps, Cohorts: cohorts, Now: now,
		LastLegacyHitAt:  now.Add(-fullrollout.LegacySilenceWindow - time.Second),
		ShadowReadSHA256: digest, LogicalDeltaSHA256: digest, ReconcileSHA256: digest,
		ProductionSnapshotSHA256: digest, V1RemovalSHA256: digest,
		BackupRestoreSHA256: digest, FrontendContractSHA256: digest,
		RollbackWindowClosureSHA256: digest,
		RollbackWindowClosedAt:      now.Add(-30 * time.Minute)}
	return fullrollout.PersistedCompletion{ReceiptID: receiptID,
		EnvironmentID: environmentID, Evidence: evidence, RecordedAt: now}
}
