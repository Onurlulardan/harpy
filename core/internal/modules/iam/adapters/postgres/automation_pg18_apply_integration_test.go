//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestIAMAutomationAppliesAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "automation_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMRBAC(t, ctx, database, origin)
			applyIAMRBACRuntime(t, ctx, database, origin)
			applyIAMAutomation(t, ctx, database, origin)
			prepareIAMAtomicRuntimeRole(t, ctx, database)
			seedIAMFoundationBase(t, ctx, database)
			for index, tenant := range iamFoundationTenants {
				seedIAMRBACCatalog(t, ctx, database, tenant.schema)
				seedIAMRBACRuntimeAuthority(t, ctx, database, tenant.schema)
				marker := "7a"
				if index == 1 {
					marker = "7b"
				}
				assertIAMServicePrincipalAudit(t, ctx, database, tenant.schema, index)
				verifyIAMAutomationFixture(t, ctx, database, tenant.schema, marker)
			}
		})
	}
}

func applyIAMAutomation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, len(iamAutomationMigrationNames))
	for _, name := range iamAutomationMigrationNames {
		documents = append(documents, iamFoundationDocument{
			path:    name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name),
		})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}

func verifyIAMAutomationFixture(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	manifestMarker string,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx, "SET LOCAL search_path = "+schema+", pg_catalog"); err != nil {
		t.Fatal(err)
	}
	fixture := strings.Replace(insertIAMAutomationFixture, "'7a'", "'"+manifestMarker+"'", 1)
	if _, err := transaction.ExecContext(ctx, fixture); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("valid automation fixture rejected: %v", err)
	}
	var revisions int
	if err := database.QueryRowContext(ctx,
		"SELECT count(*) FROM "+schema+".iam_capability_grant_revisions",
	).Scan(&revisions); err != nil || revisions != 1 {
		t.Fatalf("automation revision count = %d, err = %v", revisions, err)
	}
	if _, err := database.ExecContext(ctx,
		"UPDATE "+schema+".iam_capability_grant_revisions SET disposition='revoked'",
	); err == nil {
		t.Fatal("immutable capability revision accepted an update")
	}
	document, err := runIAMAtomicText(
		ctx, database, schema,
		`SELECT iam_capability_grant_read_v1(
            pg_catalog.current_schema()::name,
            '72000000-0000-4000-8000-000000000001', 1
        )::text`,
	)
	if err != nil || !strings.Contains(document, strings.Repeat(manifestMarker, 32)) {
		t.Fatalf("runtime evidence read failed isolation marker: %v / %s", err, document)
	}
	foreignSchema := "t_iam_a"
	if schema == foreignSchema {
		foreignSchema = "t_iam_b"
	}
	_, err = runIAMAtomicText(
		ctx, database, schema,
		`SELECT iam_capability_grant_read_v1(
            $1::name, '72000000-0000-4000-8000-000000000001', 1
        )::text`,
		foreignSchema,
	)
	if err == nil {
		t.Fatal("runtime evidence read accepted a foreign tenant schema")
	}
}

const insertIAMAutomationFixture = `
INSERT INTO iam_service_principals (
    id, principal_key, kind, display_name, state, authorization_version,
    revision, created_at, updated_at, disabled_at, retired_at
) VALUES (
    '71000000-0000-4000-8000-000000000001', 'workflow.fixture', 'workflow',
    'Workflow fixture', 'active', 1, 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL, NULL
);
INSERT INTO iam_capability_grants (
    id, principal_id, capability_key, current_revision, state,
    revision, created_at, updated_at, revoked_at
) VALUES (
    '72000000-0000-4000-8000-000000000001',
    '71000000-0000-4000-8000-000000000001',
    'forms.record.update', 1, 'active', 1,
    '2026-07-16T00:00:00Z', '2026-07-16T00:00:00Z', NULL
);
INSERT INTO iam_capability_grant_revisions (
    grant_id, grant_revision, previous_revision, principal_id, capability_key,
    scope_kind, owner_domain, resource_kind, resource_id,
    valid_from, valid_until, disposition, manifest_sha256,
    created_by_user_id, created_at
) VALUES (
    '72000000-0000-4000-8000-000000000001', 1, NULL,
    '71000000-0000-4000-8000-000000000001', 'forms.record.update',
    'resource', 'forms', 'form', '73000000-0000-4000-8000-000000000001',
    '2026-07-16T00:00:00Z', NULL, 'active',
    pg_catalog.decode(pg_catalog.repeat('7a', 32), 'hex'),
    '51000000-0000-4000-8000-000000000001', '2026-07-16T00:00:00Z'
);`
