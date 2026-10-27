//go:build integration

package integrationtest

import (
	"context"
	"testing"
)

func assertCatalogDemoProjectionCount(
	t *testing.T, ctx context.Context, database catalogPG18,
	wantCount int, wantVersion string, wantReady bool,
) {
	t.Helper()
	var count int
	var version string
	var ready bool
	err := database.owner.QueryRowContext(ctx, `SELECT count(*),
COALESCE(max(package_version),''),COALESCE(bool_and(ready),false)
FROM catalog_a.catalog_demo_provisions`).Scan(&count, &version, &ready)
	if err != nil || count != wantCount || version != wantVersion || ready != wantReady {
		t.Fatalf("Catalog demo projection=%d/%q/%t want=%d/%q/%t err=%v",
			count, version, ready, wantCount, wantVersion, wantReady, err)
	}
}

func insertCatalogDemoHandoff(
	t *testing.T, ctx context.Context, database catalogPG18, receiptID string,
	handoffID string, formsRunID string,
) {
	t.Helper()
	result, err := database.owner.ExecContext(ctx, `INSERT INTO catalog_a.catalog_demo_handoffs(
handoff_id,provision_receipt_id,component_id,package_id,package_version,
release_id,forms_run_id,content_sha256,request_sha256,
form_count,row_count,readback_sha256,completed_at)
SELECT $2::uuid,
 provision.receipt_id,provision.component_id,provision.package_id,
 provision.package_version,provision.release_id,$3::uuid,
 provision.content_sha256,provision.request_sha256,payload.form_count,payload.row_count,
 decode(repeat('ab',32),'hex'),provision.completed_at
FROM catalog_a.catalog_demo_provisions AS provision
JOIN catalog_a.catalog_demo_payloads AS payload
 ON payload.package_id=provision.package_id
AND payload.package_version=provision.package_version
AND payload.component_id=provision.component_id
AND payload.release_id=provision.release_id
WHERE provision.receipt_id=$1`, receiptID, handoffID, formsRunID)
	if err != nil {
		t.Fatalf("insert Forms handoff evidence: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		t.Fatalf("Forms handoff rows=%d err=%v", affected, err)
	}
}

func assertCatalogDemoActivationEvidence(
	t *testing.T, ctx context.Context, database catalogPG18, receiptID string,
) {
	t.Helper()
	var component, packageID, version, release, content, completed, ready bool
	err := database.owner.QueryRowContext(ctx, `SELECT
handoff.component_id=provision.component_id,
handoff.package_id=provision.package_id,
handoff.package_version=provision.package_version,
handoff.release_id=provision.release_id,
handoff.content_sha256=provision.content_sha256,
handoff.completed_at=provision.completed_at,
provision.ready AND provision.retry_at IS NULL
FROM catalog_a.catalog_demo_provisions AS provision
JOIN catalog_a.catalog_demo_handoffs AS handoff
  ON handoff.provision_receipt_id=provision.receipt_id
WHERE provision.receipt_id=$1`, receiptID).Scan(
		&component, &packageID, &version, &release, &content, &completed, &ready,
	)
	if err != nil || !component || !packageID || !version || !release ||
		!content || !completed || !ready {
		t.Fatalf("Catalog demo activation evidence=%t/%t/%t/%t/%t/%t/%t err=%v",
			component, packageID, version, release, content, completed, ready, err)
	}
}

func assertCatalogDemoActiveHead(
	t *testing.T, ctx context.Context, database catalogPG18,
	wantVersion string, wantContent string,
) {
	t.Helper()
	var componentContent, headVersion, headContent string
	err := database.owner.QueryRowContext(ctx, `SELECT
component.content_sha256,head.package_version,head.content_sha256
FROM catalog_a.catalog_catalog_components AS component
JOIN catalog_a.catalog_demo_component_heads AS head
  ON head.component_id=component.component_id
WHERE component.component_id=$1`, catalogDemoComponentID).
		Scan(&componentContent, &headVersion, &headContent)
	if err != nil || componentContent != wantContent ||
		headVersion != wantVersion || headContent != wantContent {
		t.Fatalf("Catalog demo head=%q/%q component=%q want=%q/%q err=%v",
			headVersion, headContent, componentContent,
			wantVersion, wantContent, err)
	}
}

func assertCatalogDemoEvidenceVersions(
	t *testing.T, ctx context.Context, database catalogPG18,
) {
	t.Helper()
	var payloads, provisions, handoffs, runs int
	err := database.owner.QueryRowContext(ctx, `SELECT
(SELECT count(*) FROM catalog_a.catalog_demo_payloads),
(SELECT count(*) FROM catalog_a.catalog_demo_provisions),
(SELECT count(*) FROM catalog_a.catalog_demo_handoffs),
(SELECT count(DISTINCT forms_run_id) FROM catalog_a.catalog_demo_handoffs)`).
		Scan(&payloads, &provisions, &handoffs, &runs)
	if err != nil || payloads != 2 || provisions != 2 ||
		handoffs != 2 || runs != 2 {
		t.Fatalf("Catalog demo evidence=%d/%d/%d runs=%d err=%v",
			payloads, provisions, handoffs, runs, err)
	}
}

func assertCatalogDemoRuntimeCannotForgeHandoff(
	t *testing.T, ctx context.Context, database catalogPG18,
) {
	t.Helper()
	var insertAllowed, updateAllowed, projectAllowed bool
	err := database.owner.QueryRowContext(ctx, `SELECT
pg_catalog.has_table_privilege('harpy_runtime',
 'catalog_a.catalog_demo_handoffs','INSERT'),
pg_catalog.has_table_privilege('harpy_runtime',
 'catalog_a.catalog_demo_provisions','UPDATE'),
pg_catalog.has_function_privilege('harpy_runtime',
 'catalog_a.catalog_demo_project_v1(name,uuid,uuid,text,text,uuid,uuid,bigint,text,text,text,text,bytea,integer,integer,bytea,text,bytea,timestamptz)',
 'EXECUTE')`).Scan(&insertAllowed, &updateAllowed, &projectAllowed)
	if err != nil || insertAllowed || updateAllowed || !projectAllowed {
		t.Fatalf("Catalog demo runtime ACL insert/update/project=%t/%t/%t err=%v",
			insertAllowed, updateAllowed, projectAllowed, err)
	}
}
