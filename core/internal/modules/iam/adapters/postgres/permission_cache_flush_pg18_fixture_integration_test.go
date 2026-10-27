//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

const permissionCachePlatformDecisionFixture = `
CREATE FUNCTION platform.iam_platform_authorize_decide_v1(
    p_tenant_id uuid,p_user_id uuid,p_expected_platform_clock bigint,p_capability text
)
RETURNS TABLE(allowed boolean,deployment_mode text,subject_id uuid,
    platform_clock bigint,reason text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $function$
BEGIN
    IF p_expected_platform_clock IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='40001',
            MESSAGE='Platform authorization context is stale';
    END IF;
    allowed := p_tenant_id='10000000-0000-4000-8000-000000000001'::uuid
        AND p_user_id='51000000-0000-4000-8000-000000000001'::uuid
        AND p_capability='platform.operations.permission_cache.flush';
    deployment_mode := 'onpremise'; subject_id := p_user_id;
    platform_clock := 7; reason := CASE WHEN allowed THEN 'granted' ELSE 'not_granted' END;
    RETURN NEXT;
END;
$function$;`

func preparePermissionCacheFlush(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	applyIAMFoundation(t, ctx, database, origin)
	seedIAMFoundationBase(t, ctx, database)
	if _, err := database.ExecContext(ctx, permissionCachePlatformDecisionFixture); err != nil {
		t.Fatalf("prepare platform capability fixture: %v", err)
	}
	for _, stem := range []string{
		"20260716012920_guard_iam_revision_step_v1",
		"20260716012921_guard_iam_foundation_identity_v1",
		"20260716012922_guard_iam_auth_identity_v1",
		"20260716012923_guard_iam_token_identity_v1",
		"20260716012924_guard_iam_user_security_versions_v1",
		"20260716012925_guard_iam_terminal_states_v1",
		"20260716012926_guard_iam_monotonic_counters_v1",
		"20260716013020_create_iam_resources",
	} {
		mustExecIAMTenant(t, ctx, database, "t_iam_a",
			readPermissionCacheMigration(t, origin, stem))
	}
	authorizationState := readPermissionCacheMigration(t, origin,
		"20260716013130_create_iam_authorization_state")
	mustExecIAMTenant(t, ctx, database, "t_iam_a", authorizationState)
	mustExecIAMTenant(t, ctx, database, "t_iam_a", `
UPDATE iam_users SET authorization_version=1,revision=revision+1,
    updated_at='2026-07-16T00:01:00Z'
WHERE id IN (
    '51000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000002'
);
INSERT INTO iam_authorization_clock(singleton,clock_value,revision,created_at,updated_at)
VALUES(true,1,1,'2026-07-16T00:01:00Z','2026-07-16T00:01:00Z');`)
	flush := readPermissionCacheMigration(t, origin,
		"20260723061024_create_iam_permission_cache_flush_v1")
	mustExecIAMTenant(t, ctx, database, "t_iam_a", flush)
	if _, err := database.ExecContext(ctx,
		"GRANT USAGE ON SCHEMA t_iam_a TO harpy_runtime"); err != nil {
		t.Fatalf("grant permission-cache schema usage: %v", err)
	}
}

func readPermissionCacheMigration(t *testing.T, origin, stem string) string {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	suffix := ".fresh.sql"
	path := filepath.Join(directory, "..", "..", "migrations", "tenant",
		origin, stem+suffix)
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read permission-cache migration %s: %v", path, err)
	}
	return string(document)
}
