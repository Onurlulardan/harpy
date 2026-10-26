//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestWorkPG18RESTQueryVisibilityRevisionAndACL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	database := startWorkPG18(t, ctx)
	prepareWorkSchemas(t, ctx, database.owner)
	seedWorkQueryFixture(t, ctx, database.owner, "work_fresh_a")
	assertWorkQueryOwnerGuard(t, ctx, database.owner)
	prepareWorkQueryRuntime(t, ctx, database.owner)
	connection, err := database.owner.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	if _, err = connection.ExecContext(ctx, `
SET SESSION AUTHORIZATION harpy_v2_runtime;
SET ROLE harpy_runtime;
SET search_path TO work_fresh_a,pg_catalog`); err != nil {
		t.Fatal(err)
	}
	assertWorkQueryRuntimeBoundary(t, ctx, connection)
	assertWorkQueryACL(t, ctx, connection)
	assertWorkQueryCounts(t, ctx, connection, workQueryPrincipal, []int{1, 1, 1, 3, 2, 4, 3})
	assertWorkQueryCounts(t, ctx, connection, workOtherPrincipal, []int{0, 0, 0, 0, 0, 0, 0})
}

func prepareWorkQueryRuntime(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
CREATE ROLE harpy_v2_runtime NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB
NOCREATEROLE NOREPLICATION NOBYPASSRLS;
GRANT harpy_runtime TO harpy_v2_runtime
WITH ADMIN FALSE, INHERIT FALSE, SET TRUE;
GRANT USAGE ON SCHEMA work_fresh_a TO harpy_runtime`)
	if err != nil {
		t.Fatal(err)
	}
}

func assertWorkQueryOwnerGuard(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	var count int
	err := database.QueryRowContext(ctx,
		"SELECT count(*) FROM work_fresh_a.work_board_query_v1").Scan(&count)
	if err != nil || count != 0 {
		t.Fatalf("owner query guard count=%d error=%v", count, err)
	}
}

func assertWorkQueryRuntimeBoundary(t *testing.T, ctx context.Context, connection *sql.Conn) {
	t.Helper()
	var session, current string
	if err := connection.QueryRowContext(ctx,
		"SELECT session_user,current_user").Scan(&session, &current); err != nil {
		t.Fatal(err)
	}
	if session != "harpy_v2_runtime" || current != "harpy_runtime" {
		t.Fatalf("runtime boundary session=%s current=%s", session, current)
	}
}

func assertWorkQueryACL(t *testing.T, ctx context.Context, connection *sql.Conn) {
	t.Helper()
	var selectView, insertView bool
	err := connection.QueryRowContext(ctx, `
SELECT has_table_privilege(current_user,'work_board_query_v1','SELECT'),
       has_table_privilege(current_user,'work_board_query_v1','INSERT')`).Scan(
		&selectView, &insertView,
	)
	if err != nil || !selectView || insertView {
		t.Fatalf("query view ACL select=%t insert=%t error=%v", selectView, insertView, err)
	}
	for _, grant := range []struct{ table, column, privilege string }{
		{"work_tasks", "creator_user_id", "INSERT"},
		{"work_tasks", "priority", "UPDATE"},
		{"work_tasks", "starts_at", "UPDATE"},
		{"work_board_columns", "color", "INSERT"},
		{"work_calendar_events", "color", "UPDATE"},
		{"work_inbox_items", "visible_to_principals", "INSERT"},
		{"work_calendar_items", "visible_to_principals", "UPDATE"},
	} {
		var allowed bool
		err = connection.QueryRowContext(ctx,
			"SELECT has_column_privilege(current_user,$1,$2,$3)",
			grant.table, grant.column, grant.privilege).Scan(&allowed)
		if err != nil || !allowed {
			t.Fatalf("missing %s(%s) %s: allowed=%t error=%v",
				grant.table, grant.column, grant.privilege, allowed, err)
		}
	}
}

func assertWorkQueryCounts(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	principal string, expected []int,
) {
	t.Helper()
	queries := []string{
		`SELECT count(*) FROM work_board_query_v1
WHERE visibility_principals && ARRAY[$1]::text[]`,
		`SELECT count(*) FROM work_task_query_v1
WHERE visibility_principals && ARRAY[$1]::text[]`,
		`SELECT count(*) FROM work_calendar_event_query_v1
WHERE visibility_principals && ARRAY[$1]::text[]`,
		`SELECT count(*) FROM work_inbox_query_v1
WHERE visibility_principals && ARRAY[$1]::text[]
  AND context_principals && ARRAY[$1]::text[]`,
		`SELECT count(*) FROM work_inbox_query_v1
WHERE visibility_principals && ARRAY[$1]::text[]
  AND context_principals && ARRAY[$1]::text[]
  AND (source_owner='WORK' OR permission_revision=7)`,
		`SELECT count(*) FROM work_calendar_query_v1
WHERE visibility_principals && ARRAY[$1]::text[]
  AND context_principals && ARRAY[$1]::text[]`,
		`SELECT count(*) FROM work_calendar_query_v1
WHERE visibility_principals && ARRAY[$1]::text[]
  AND context_principals && ARRAY[$1]::text[]
  AND (source_owner='WORK' OR permission_revision=7)`,
	}
	if len(expected) != len(queries) {
		t.Fatalf("expected count vector length=%d, want %d", len(expected), len(queries))
	}
	for index, query := range queries {
		var count int
		if err := connection.QueryRowContext(ctx, query, principal).Scan(&count); err != nil {
			t.Fatalf("query %d for %s: %v", index, principal, err)
		}
		if count != expected[index] {
			t.Fatalf("query %d for %s count=%d want=%d",
				index, principal, count, expected[index])
		}
	}
}
