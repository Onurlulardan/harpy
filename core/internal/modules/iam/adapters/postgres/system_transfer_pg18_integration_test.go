//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestIAMSystemTransferAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "system_transfer_"+origin)
			prepareIAMFormsRecordAuthority(t, ctx, database, origin)
			applyIAMSystemTransferAsyncPlatform(t, ctx, database, origin)
			grantIAMSystemTransferSchemaUsage(t, ctx, database)
			applyIAMSystemTransferDependencies(t, ctx, database, origin)
			applyIAMSystemTransfer(t, ctx, database, origin)
			assertIAMSystemTransferCatalog(t, ctx, database)
			assertIAMExportReconciliationExpiredRecovery(t, ctx, database)
			assertIAMSystemImportSourceReplay(t, ctx, database)
			assertIAMSystemTransferIsolation(t, ctx, database)
			assertIAMSystemTransferCapacity(t, ctx, database)
		})
	}
}

func applyIAMSystemTransferDependencies(t *testing.T, ctx context.Context,
	database *sql.DB, origin string) {
	t.Helper()
	document := iamFoundationDocument{
		path: "20260716013700_create_iam_identity_authority_v1." + origin + ".sql",
		content: readIAMIdentityMigration(t, origin,
			"20260716013700_create_iam_identity_authority_v1"),
	}
	for _, tenant := range iamFoundationTenants {
		sentinelOrigin := "new"
		if origin == "v1_adopted" {
			sentinelOrigin = origin
		}
		mustExecIAMTenant(t, ctx, database, tenant.schema, `
CREATE TABLE tenancy_schema_identity(
 singleton boolean PRIMARY KEY CHECK(singleton),tenant_id uuid NOT NULL UNIQUE,
 schema_name text NOT NULL,origin text NOT NULL,
 created_at timestamptz NOT NULL
		)`)
		mustExecIAMTenant(t, ctx, database, tenant.schema, `
INSERT INTO tenancy_schema_identity VALUES(true,$1,$2,$3,'2026-07-18T00:00:00Z')`,
			tenant.id, tenant.schema, sentinelOrigin)
		applyIAMTenantDocuments(t, ctx, database, tenant.schema,
			[]iamFoundationDocument{document})
		mustExecIAMTenant(t, ctx, database, tenant.schema, `UPDATE iam_users
SET authorization_version=1,revision=revision+1,
    updated_at='2026-07-18T00:00:00Z'
WHERE id='51000000-0000-4000-8000-000000000001'`)
	}
}

func grantIAMSystemTransferSchemaUsage(t *testing.T, ctx context.Context,
	database *sql.DB) {
	t.Helper()
	if _, err := database.ExecContext(ctx,
		"GRANT USAGE ON SCHEMA t_iam_a,t_iam_b TO harpy_runtime"); err != nil {
		t.Fatalf("grant IAM transfer schema usage: %v", err)
	}
}

func applyIAMSystemTransfer(t *testing.T, ctx context.Context,
	database *sql.DB, origin string) {
	t.Helper()
	documents := make([]iamFoundationDocument, 0, len(iamSystemTransferMigrationNames))
	for _, name := range iamSystemTransferMigrationNames {
		documents = append(documents, iamFoundationDocument{path: name + "." + origin + ".sql",
			content: readIAMIdentityMigration(t, origin, name)})
	}
	for _, tenant := range iamFoundationTenants {
		applyIAMTenantDocuments(t, ctx, database, tenant.schema, documents)
	}
}

func assertIAMSystemTransferCatalog(t *testing.T, ctx context.Context,
	database *sql.DB) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		connection, transaction := beginIAMTenantTx(t, ctx, database, tenant.schema)
		var permissions, tables, claimFunctions int
		var claimExecutable bool
		err := transaction.QueryRowContext(ctx, `SELECT
(SELECT pg_catalog.count(*) FROM iam_permissions permission
 JOIN iam_resources resource ON resource.id=permission.resource_id
 JOIN iam_actions action ON action.id=permission.action_id
 WHERE resource.resource_key IN ('iam.users','iam.roles','iam.organizations')
 AND action.action_key IN ('import','export') AND permission.match_kind='exact'
 AND permission.is_protected AND permission.state='active'),
(SELECT pg_catalog.count(*) FROM pg_catalog.pg_class
 WHERE relnamespace=pg_catalog.current_schema()::regnamespace
 AND relname LIKE 'iam_system_%' AND relkind='r'),
(SELECT pg_catalog.count(*) FROM pg_catalog.pg_proc
 WHERE pronamespace=pg_catalog.current_schema()::regnamespace
 AND proname='iam_system_record_import_worker_claim_v2'),
COALESCE((SELECT pg_catalog.has_function_privilege(
 'harpy_runtime',routine.oid,'EXECUTE') FROM pg_catalog.pg_proc AS routine
 WHERE routine.pronamespace=pg_catalog.current_schema()::regnamespace
 AND routine.proname='iam_system_record_import_worker_claim_v2'),false)`).Scan(
			&permissions, &tables, &claimFunctions, &claimExecutable)
		transaction.Rollback()
		connection.Close()
		if err != nil || permissions != 6 || tables != 5 || claimFunctions != 1 ||
			!claimExecutable {
			t.Fatalf("tenant %s transfer catalog=%d/%d claim=%d/%t: %v",
				tenant.schema, permissions, tables, claimFunctions, claimExecutable, err)
		}
	}
}

func assertIAMSystemTransferIsolation(t *testing.T, ctx context.Context,
	database *sql.DB) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, iamFoundationTenants[0].schema, `
INSERT INTO iam_system_record_exports (
 export_id,tenant_id,job_id,idempotency_receipt_id,resource_kind,actor_id,
 token_version,authorization_version,authorization_clock,catalog_version_id,
 catalog_sha256,query_document,query_sha256,request_sha256,
 idempotency_key_sha256,format,created_at,metadata_retain_until
) VALUES (
 '019be000-0000-7000-8000-000000000001','10000000-0000-4000-8000-000000000001',
 '019be000-0000-7000-8000-000000000002','019be000-0000-7000-8000-000000000003',
 'users','91000000-0000-4000-8000-000000000001',1,1,1,
 '019be000-0000-7000-8000-000000000004',decode(repeat('11',32),'hex'),
 convert_to('{}','UTF8'),sha256(convert_to('{}','UTF8')),decode(repeat('22',32),'hex'),
 decode(repeat('33',32),'hex'),'csv','2026-07-18T00:00:00Z','2026-10-18T00:00:00Z'
)`)
	connection, transaction := beginIAMTenantTx(t, ctx, database,
		iamFoundationTenants[1].schema)
	defer connection.Close()
	defer transaction.Rollback()
	var count int
	if err := transaction.QueryRowContext(ctx,
		`SELECT pg_catalog.count(*) FROM iam_system_record_exports`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("foreign tenant transfer rows=%d: %v", count, err)
	}
}

func assertIAMSystemTransferCapacity(t *testing.T, ctx context.Context,
	database *sql.DB) {
	t.Helper()
	const twoSlots = `INSERT INTO iam_system_transfer_admissions
(transfer_kind,transfer_id,tenant_id,actor_id,state,tenant_run_slot,admitted_at)
VALUES
('export','019be000-0000-7000-8000-000000000011','10000000-0000-4000-8000-000000000001',
 '91000000-0000-4000-8000-000000000001','running',1,'2026-07-18T00:00:00Z'),
('import','019be000-0000-7000-8000-000000000012','10000000-0000-4000-8000-000000000001',
 '91000000-0000-4000-8000-000000000002','running',2,'2026-07-18T00:00:00Z')`
	mustExecIAMTenant(t, ctx, database, iamFoundationTenants[0].schema, twoSlots)
	connection, transaction := beginIAMTenantTx(t, ctx, database,
		iamFoundationTenants[0].schema)
	defer connection.Close()
	defer transaction.Rollback()
	_, err := transaction.ExecContext(ctx, `INSERT INTO iam_system_transfer_admissions
(transfer_kind,transfer_id,tenant_id,actor_id,state,tenant_run_slot,admitted_at)
VALUES('export','019be000-0000-7000-8000-000000000013',
'10000000-0000-4000-8000-000000000001','91000000-0000-4000-8000-000000000003',
'running',2,'2026-07-18T00:00:00Z')`)
	if err == nil {
		t.Fatal("third tenant transfer reused a running slot")
	}
}
