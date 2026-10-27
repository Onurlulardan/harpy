//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func assertAsyncACLPrivileges(
	t *testing.T, ctx context.Context, runtime *pgxpool.Conn, database *sql.DB,
) {
	t.Helper()
	for _, table := range []string{
		"platform.jobs", "platform.job_attempts", "platform.outbox",
		"platform.inbox", "platform.idempotency_receipts", "platform.dead_letters",
	} {
		assertACLTablePrivilege(t, ctx, database, "harpy_runtime", table, "SELECT", true)
		assertACLTablePrivilege(t, ctx, database, "harpy_runtime", table, "INSERT", false)
		assertACLTablePrivilege(t, ctx, database, "harpy_runtime", table, "DELETE", false)
		assertACLTablePrivilege(t, ctx, database, "harpy_v2_runtime", table, "SELECT", false)
		assertACLTablePrivilege(t, ctx, database, "harpy_adopter", table, "SELECT", false)
		assertACLTablePrivilege(t, ctx, database, "harpy_cluster_admin", table, "INSERT", false)
	}
	for _, table := range []string{
		"platform.cutover_groups", "platform.cutover_group_members",
		"platform.domain_cutovers",
	} {
		assertACLTablePrivilege(t, ctx, database, "harpy_runtime", table, "SELECT", true)
		assertACLTablePrivilege(t, ctx, database, "harpy_runtime", table, "UPDATE", false)
	}
	for _, column := range []string{
		"state", "attempts", "lease_generation", "lease_owner_id",
		"lease_heartbeat_at", "lease_until", "cancel_requested_at",
		"effect_started_at", "run_at", "last_error_class", "last_error_code",
		"last_error_sha256", "retain_until", "updated_at", "completed_at",
	} {
		assertACLColumnPrivilege(t, ctx, database, "platform.jobs", column, "UPDATE", true)
	}
	for _, column := range []string{"kind", "payload_ref", "parked_at"} {
		assertACLColumnPrivilege(t, ctx, database, "platform.jobs", column, "UPDATE", false)
	}
	for _, column := range []string{
		"finished_at", "outcome", "error_class", "error_code",
		"redacted_error_sha256", "effect_started_at", "effect_id",
		"effect_request_sha256", "effect_receipt_id",
	} {
		assertACLColumnPrivilege(t, ctx, database, "platform.job_attempts", column, "UPDATE", true)
	}
	for _, column := range []string{"attempt", "job_id", "started_at"} {
		assertACLColumnPrivilege(t, ctx, database, "platform.job_attempts", column, "UPDATE", false)
	}
	for _, column := range []string{
		"delivery_state", "delivery_attempts", "lease_owner_id", "lease_generation",
		"lease_heartbeat_at", "lease_until", "next_attempt_at", "confirm_sha256",
		"last_error_code", "last_error_sha256", "updated_at",
	} {
		assertACLColumnPrivilege(t, ctx, database, "platform.outbox", column, "UPDATE", true)
	}
	for _, column := range []string{
		"event_type", "payload_canonical", "confirmed_at", "delivered_at",
		"parked_at", "parked_code", "parked_current_group_generation",
		"parked_current_cutover_generation",
	} {
		assertACLColumnPrivilege(t, ctx, database, "platform.outbox", column, "UPDATE", false)
	}
	assertACLInsertColumns(t, ctx, database)
	for _, function := range []string{
		"platform.platform_async_assert_current_authority_v1(text,uuid,text,bigint,text,bigint)",
		"platform.platform_async_dispose_stale_job_v1(uuid,text,uuid,uuid,uuid,uuid,text,uuid,text,bigint,text,bigint,text,text,bytea,text,text,text,uuid,uuid,text,text,text)",
		"platform.platform_async_event_gap_job_ready_v2(uuid)",
		"platform.platform_async_park_stale_outbox_v1(uuid,uuid)",
		"platform.platform_async_park_stale_work_v1(uuid,uuid,bigint,uuid,bytea,text,jsonb,text)",
		"platform.platform_async_park_stale_work_v2(uuid,text,uuid,bigint,uuid)",
		"platform.platform_async_reconcile_event_gap_v2(uuid,uuid,uuid,bigint)",
		"platform.platform_async_release_job_retention_v1(uuid,uuid,text,bigint,text,bigint,bigint,text,timestamptz)",
		"platform.platform_async_run_retention_cycle_v1()",
		"platform.platform_async_validate_job_attempt_v1(platform.job_attempts,platform.job_attempts,name,name)",
		"platform.platform_async_validate_job_transition_v1(platform.jobs,platform.jobs,name,name)",
	} {
		assertACLFunctionPrivilege(t, ctx, database, "harpy_runtime", function, true)
	}
	for _, function := range []string{
		"platform.platform_async_event_gap_job_ready_v1(uuid)",
		"platform.platform_async_purge_dead_letters_v2(integer)",
		"platform.platform_async_purge_idempotency_receipts_v1(integer)",
		"platform.platform_async_purge_inbox_v1(integer)",
		"platform.platform_async_purge_jobs_v1(integer)",
		"platform.platform_async_outbox_authority_current_v1(platform.outbox)",
		"platform.platform_async_reconcile_event_gap_v1(uuid,uuid,uuid,bigint)",
		"platform.platform_async_record_stale_job_disposition_v1(platform.jobs,uuid,uuid,uuid,uuid,text,text,bigint,bigint,text,bytea,text,text,text,uuid,uuid,text,text,text,timestamptz)",
		"platform.platform_async_redact_dead_letter_payloads_v1(integer)",
		"platform.platform_async_redact_dead_letter_payloads_v2(integer)",
		"platform.platform_async_redact_job_payloads_v1(integer)",
		"platform.platform_async_redact_outbox_payloads_v1(integer)",
		"platform.platform_async_redact_outbox_payloads_v2(integer)",
		"platform.platform_async_reissue_stale_event_gap_v1(platform.jobs,uuid,bigint,bigint,timestamptz)",
		"platform.platform_async_validate_stale_job_disposition_v1(platform.jobs,uuid,uuid,uuid,uuid,text,text,bigint,bigint,text,bytea,text,text,text,uuid,uuid,text,text,text,timestamptz)",
		"platform.platform_async_verify_stale_job_disposition_replay_v1(platform.jobs,uuid,uuid,uuid,bigint,bigint,text,text,text,uuid,uuid)",
		"platform.platform_async_guard_job_transition_v1()",
		"platform.platform_async_guard_job_transition_v2()",
		"platform.platform_async_guard_job_attempt_v2()",
		"platform.platform_async_guard_outbox_park_v1()",
	} {
		assertACLFunctionPrivilege(t, ctx, database, "harpy_runtime", function, false)
	}
	assertACLAdoptionFunctions(t, ctx, database)
	assertACLSchemaPrivilege(t, ctx, database, "harpy_runtime", true)
	assertACLSchemaPrivilege(t, ctx, database, "harpy_v2_runtime", false)
	var current, session string
	if err := runtime.QueryRow(ctx, "SELECT current_user, session_user").Scan(&current, &session); err != nil {
		t.Fatalf("read runtime identities: %v", err)
	}
	if current != "harpy_runtime" || session != "harpy_v2_runtime" {
		t.Fatalf("runtime identities = %s/%s", current, session)
	}
}

func assertACLTablePrivilege(
	t *testing.T, ctx context.Context, database *sql.DB,
	role, table, privilege string, want bool,
) {
	t.Helper()
	var got bool
	if err := database.QueryRowContext(ctx,
		"SELECT pg_catalog.has_table_privilege($1,$2,$3)",
		role, table, privilege).Scan(&got); err != nil {
		t.Fatalf("inspect %s %s on %s: %v", role, privilege, table, err)
	}
	if got != want {
		t.Fatalf("%s %s on %s = %t, want %t", role, privilege, table, got, want)
	}
}

func assertACLColumnPrivilege(
	t *testing.T, ctx context.Context, database *sql.DB,
	table, column, privilege string, want bool,
) {
	t.Helper()
	var got bool
	if err := database.QueryRowContext(ctx,
		"SELECT pg_catalog.has_column_privilege('harpy_runtime',$1,$2,$3)",
		table, column, privilege).Scan(&got); err != nil {
		t.Fatalf("inspect %s on %s.%s: %v", privilege, table, column, err)
	}
	if got != want {
		t.Fatalf("runtime %s on %s.%s = %t, want %t",
			privilege, table, column, got, want)
	}
}

func assertACLFunctionPrivilege(
	t *testing.T, ctx context.Context, database *sql.DB,
	role, function string, want bool,
) {
	t.Helper()
	var got bool
	if err := database.QueryRowContext(ctx,
		"SELECT pg_catalog.has_function_privilege($1,$2,'EXECUTE')",
		role, function).Scan(&got); err != nil {
		t.Fatalf("inspect %s EXECUTE on %s: %v", role, function, err)
	}
	if got != want {
		t.Fatalf("%s EXECUTE on %s = %t, want %t", role, function, got, want)
	}
}

func assertACLSchemaPrivilege(
	t *testing.T, ctx context.Context, database *sql.DB, role string, want bool,
) {
	t.Helper()
	var got bool
	if err := database.QueryRowContext(ctx,
		"SELECT pg_catalog.has_schema_privilege($1,'platform','USAGE')", role,
	).Scan(&got); err != nil {
		t.Fatalf("inspect %s schema USAGE: %v", role, err)
	}
	if got != want {
		t.Fatalf("%s platform schema USAGE = %t, want %t", role, got, want)
	}
}
