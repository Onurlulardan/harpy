//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func seedAnalyticsRetention(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	seedAnalyticsRetiredDraft(t, ctx, database,
		"019b1000-0000-7000-8000-000000000101", retentionOldDraft, "31 days")
	seedAnalyticsRetiredDraft(t, ctx, database,
		"019b1000-0000-7000-8000-000000000103", retentionRecentDraft, "29 days")
	seedAnalyticsPublishedVersion(t, ctx, database,
		"019b1000-0000-7000-8000-000000000105", retentionHeldVersion)
	seedAnalyticsRetentionAudits(t, ctx, database)
}

func seedAnalyticsRetiredDraft(
	t *testing.T, ctx context.Context, database *sql.DB,
	rootID, versionID, age string,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx, `SET CONSTRAINTS ALL DEFERRED`); err != nil {
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx, `INSERT INTO t_analytics_a.analytics_saved_queries (
 id,lifecycle_state,row_revision,reference_epoch,current_draft_version_id,
 created_by,updated_by,created_at,updated_at
) VALUES ($1,'retired',1,1,$2,$3,$3,clock_timestamp()-$4::interval,
clock_timestamp()-$4::interval)`, rootID, versionID, retentionActor, age)
	if err == nil {
		_, err = transaction.ExecContext(ctx, `INSERT INTO t_analytics_a.analytics_saved_query_versions (
 id,query_id,version_no,version_state,row_revision,name,query_document,query_sha256,
 parameter_document,parameter_sha256,output_document,output_sha256,manifest_sha256,
 cache_policy,created_by,updated_by,created_at,updated_at
) VALUES ($2,$1,1,'draft',1,'retention draft','{}'::bytea,sha256('{}'::bytea),
'{}'::bytea,sha256('{}'::bytea),'{}'::bytea,sha256('{}'::bytea),
decode(repeat('61',32),'hex'),'none',$3,$3,
clock_timestamp()-$4::interval,clock_timestamp()-$4::interval)`,
			rootID, versionID, retentionActor, age)
	}
	if err != nil {
		t.Fatalf("seed Analytics retired draft: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func seedAnalyticsPublishedVersion(
	t *testing.T, ctx context.Context, database *sql.DB, rootID, versionID string,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx, `SET CONSTRAINTS ALL DEFERRED`); err != nil {
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx, `INSERT INTO t_analytics_a.analytics_saved_queries (
id,lifecycle_state,row_revision,reference_epoch,current_draft_version_id,
created_by,updated_by,created_at,updated_at
) VALUES ($1,'active',1,1,$3,$2,$2,clock_timestamp()-interval '500 days',
clock_timestamp()-interval '500 days')`, rootID, retentionActor, retentionHeldDraft)
	if err == nil {
		_, err = transaction.ExecContext(ctx, `INSERT INTO t_analytics_a.analytics_saved_query_versions (
id,query_id,version_no,version_state,row_revision,name,query_document,query_sha256,
parameter_document,parameter_sha256,output_document,output_sha256,manifest_sha256,
cache_policy,created_by,updated_by,created_at,updated_at,published_by,published_at
) VALUES ($2,$1,1,'published',1,'referenced query','{}'::bytea,sha256('{}'::bytea),
'{}'::bytea,sha256('{}'::bytea),'{}'::bytea,sha256('{}'::bytea),
decode(repeat('62',32),'hex'),'none',$3,$3,clock_timestamp()-interval '500 days',
clock_timestamp()-interval '500 days',$3,clock_timestamp()-interval '499 days')`,
			rootID, versionID, retentionActor)
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, `INSERT INTO t_analytics_a.analytics_saved_query_versions (
id,query_id,version_no,version_state,row_revision,name,query_document,query_sha256,
parameter_document,parameter_sha256,output_document,output_sha256,manifest_sha256,
cache_policy,created_by,updated_by,created_at,updated_at,published_by,published_at
) VALUES ($2,$1,2,'draft',1,'referenced query next draft','{}'::bytea,
sha256('{}'::bytea),'{}'::bytea,sha256('{}'::bytea),'{}'::bytea,
sha256('{}'::bytea),decode(repeat('63',32),'hex'),'none',$3,$3,
clock_timestamp()-interval '450 days',clock_timestamp()-interval '401 days',NULL,NULL)`,
			rootID, retentionHeldDraft, retentionActor)
	}
	if err != nil {
		t.Fatalf("seed referenced Analytics version: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatalf("commit referenced Analytics version: %v", err)
	}
}
