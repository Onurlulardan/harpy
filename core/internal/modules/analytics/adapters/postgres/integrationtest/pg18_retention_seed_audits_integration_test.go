//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func seedAnalyticsRetentionAudits(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	execAnalyticsRetentionSeed(t, ctx, database, `INSERT INTO t_analytics_a.analytics_publication_releases (
id,resource_kind,resource_version_id,audience,lifecycle_state,cache_policy,
manifest_sha256,requester_id,row_revision,generation,created_at,updated_at,
activated_at,revoked_at
) VALUES ('019b1000-0000-7000-8000-000000000120','query',$1,'tenant','revoked','none',
decode(repeat('70',32),'hex'),$2,1,1,clock_timestamp()-interval '500 days',
clock_timestamp()-interval '399 days',clock_timestamp()-interval '450 days',
clock_timestamp()-interval '399 days')`,
		"019b1000-0000-7000-8000-000000000130", retentionActor)
	execAnalyticsRetentionSeed(t, ctx, database, `INSERT INTO t_analytics_a.analytics_execution_audits
VALUES
('019b1000-0000-7000-8000-000000000121','019b1000-0000-7000-8000-000000000120',
'actor',decode(repeat('71',32),'hex'),decode(repeat('72',32),'hex'),
decode(repeat('73',32),'hex'),1,false,clock_timestamp()-interval '401 days'),
('019b1000-0000-7000-8000-000000000122','019b1000-0000-7000-8000-000000000120',
'actor',decode(repeat('74',32),'hex'),decode(repeat('75',32),'hex'),
decode(repeat('76',32),'hex'),1,false,clock_timestamp()-interval '399 days')`)
	execAnalyticsRetentionSeed(t, ctx, database, `INSERT INTO t_analytics_a.analytics_publication_approvals
VALUES
('019b1000-0000-7000-8000-000000000123',$1,
clock_timestamp()-interval '7 years 1 day',$2,
clock_timestamp()-interval '7 years 1 day',clock_timestamp()-interval '7 years 1 day',
decode(repeat('77',32),'hex')),
('019b1000-0000-7000-8000-000000000124',$1,
clock_timestamp()-interval '7 years'+interval '1 day',$2,
clock_timestamp()-interval '7 years'+interval '1 day',
clock_timestamp()-interval '7 years'+interval '1 day',decode(repeat('78',32),'hex'))`,
		retentionActor, "019b1000-0000-7000-8000-000000000127")
	execAnalyticsRetentionSeed(t, ctx, database, `INSERT INTO t_analytics_a.analytics_operations VALUES
('019b1000-0000-7000-8000-000000000125',$1,'query_delete',NULL,
decode(repeat('79',32),'hex'),'query',$2,1,decode(repeat('7a',32),'hex'),
clock_timestamp()-interval '7 years 1 day'),
('019b1000-0000-7000-8000-000000000126',$1,'query_delete',NULL,
decode(repeat('7b',32),'hex'),'query',$2,1,decode(repeat('7c',32),'hex'),
clock_timestamp()-interval '7 years'+interval '1 day')`, retentionActor,
		retentionHeldVersion)
	seedAnalyticsCapabilities(t, ctx, database)
}

func seedAnalyticsCapabilities(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	execAnalyticsRetentionSeed(t, ctx, database, `INSERT INTO t_analytics_a.analytics_publication_releases (
id,resource_kind,resource_version_id,audience,lifecycle_state,cache_policy,
manifest_sha256,requester_id,approver_id,row_revision,generation,public_id,
capability_expires_at,created_at,updated_at,activated_at,revoked_at
) VALUES
('019b1000-0000-7000-8000-000000000128','query',$1,'capability','revoked','none',
decode(repeat('80',32),'hex'),$2,$3,1,1,decode(repeat('81',16),'hex'),
clock_timestamp()-interval '31 days',clock_timestamp()-interval '70 days',
clock_timestamp()-interval '31 days',clock_timestamp()-interval '60 days',
clock_timestamp()-interval '31 days'),
('019b1000-0000-7000-8000-000000000129','query',$4,'capability','revoked','none',
decode(repeat('82',32),'hex'),$2,$3,1,1,decode(repeat('83',16),'hex'),
clock_timestamp()-interval '29 days',clock_timestamp()-interval '68 days',
clock_timestamp()-interval '29 days',clock_timestamp()-interval '58 days',
clock_timestamp()-interval '29 days')`,
		"019b1000-0000-7000-8000-000000000131", retentionActor,
		"019b1000-0000-7000-8000-000000000127",
		"019b1000-0000-7000-8000-000000000132")
	execAnalyticsRetentionSeed(t, ctx, database, `INSERT INTO t_analytics_a.analytics_capability_tokens VALUES
('019b1000-0000-7000-8000-000000000128',decode(repeat('84',32),'hex'),
clock_timestamp()-interval '60 days',clock_timestamp()-interval '31 days',
clock_timestamp()-interval '31 days'),
('019b1000-0000-7000-8000-000000000129',decode(repeat('85',32),'hex'),
clock_timestamp()-interval '58 days',clock_timestamp()-interval '29 days',
clock_timestamp()-interval '29 days')`)
}

func execAnalyticsRetentionSeed(
	t *testing.T, ctx context.Context, database *sql.DB,
	statement string, arguments ...any,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx, statement, arguments...); err != nil {
		t.Fatalf("seed Analytics retention evidence: %v", err)
	}
}
