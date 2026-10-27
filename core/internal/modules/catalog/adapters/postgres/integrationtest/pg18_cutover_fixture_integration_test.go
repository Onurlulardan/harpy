//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

const catalogCutoverInstallationID = "018f2000-0000-7000-8000-000000000070"
const catalogCutoverRoute = "api.modules.v1"
const catalogCutoverRetiredAt = "2026-07-21T01:00:00Z"

type catalogCutoverMapping struct {
	checkpoint, sourceIdentity, sourceDocument string
	targetKind, targetIdentity, targetRecord   string
}

func loadCatalogCutoverMappings(t *testing.T, ctx context.Context,
	database *sql.DB, authority catalogCutoverAuthority,
) []catalogCutoverMapping {
	t.Helper()
	moduleDocument := queryCatalogCutoverText(t, ctx, database, `SELECT
to_jsonb(row)::text FROM catalog_a.modules AS row WHERE row.id=$1`, catalogPackageID)
	versionID := catalogLegacySchemaVersionID(catalogTenantA)
	versionDocument := queryCatalogCutoverText(t, ctx, database, `SELECT
to_jsonb(row)::text FROM catalog_a.module_schema_versions AS row WHERE row.id=$1`,
		versionID)
	installation := queryCatalogCutoverText(t, ctx, database, `SELECT
pg_catalog.jsonb_build_object(
'installation_id',$1::uuid,'package_id',$2::uuid,'active_version','1.0.0',
'state','ACTIVE','reference_epoch',1,'row_version',1,
'updated_at',$3::timestamptz)::text`, catalogCutoverInstallationID,
		catalogPackageID, "2026-07-21T10:00:00Z")
	binding := queryCatalogCutoverText(t, ctx, database, `SELECT
pg_catalog.jsonb_build_object(
'installation_id',$1::uuid,'component_id',$2::uuid,'owner_module','catalog',
'owner_generation',1,'owner_contract','catalog-evidence.v1',
'content_sha256',$3::text,'state','ACTIVE')::text`,
		catalogCutoverInstallationID, catalogComponentID, catalogManifestDigest)
	result := []catalogCutoverMapping{
		{checkpoint: "modules", sourceIdentity: catalogPackageID,
			sourceDocument: moduleDocument, targetKind: "installation",
			targetIdentity: catalogCutoverInstallationID, targetRecord: installation},
		{checkpoint: "module_schema_versions", sourceIdentity: versionID,
			sourceDocument: versionDocument, targetKind: "binding",
			targetIdentity: catalogComponentID, targetRecord: binding},
	}
	return append(result, loadCatalogCutoverMenuMappings(t, ctx, database)...)
}

func seedCatalogCutoverConsumer(t *testing.T, ctx context.Context,
	database *sql.DB,
) []catalogRootEntry {
	t.Helper()
	execCatalogCutoverSeed(t, ctx, database, `INSERT INTO catalog_a.catalog_legacy_route_hits(
route_key,hit_count,last_hit_at,consumer_ack_count,required_consumer_count,
remove_after,retired_at) VALUES($1,1,'2026-07-20T23:00:00Z',1,1,
'2026-07-21T00:00:00Z',$2::timestamptz)`, catalogCutoverRoute,
		catalogCutoverRetiredAt)
	execCatalogCutoverSeed(t, ctx, database, `INSERT INTO catalog_a.catalog_legacy_consumer_acks(
route_key,consumer_id,evidence_sha256,acknowledged_at)
VALUES($1,'frontend-v3',$2,'2026-07-20T23:30:00Z')`, catalogCutoverRoute,
		"7171717171717171717171717171717171717171717171717171717171717171")
	route := queryCatalogCutoverText(t, ctx, database, `SELECT to_jsonb(row)::text
FROM catalog_a.catalog_legacy_route_hits AS row WHERE row.route_key=$1`,
		catalogCutoverRoute)
	ack := queryCatalogCutoverText(t, ctx, database, `SELECT to_jsonb(row)::text
FROM catalog_a.catalog_legacy_consumer_acks AS row WHERE row.route_key=$1`,
		catalogCutoverRoute)
	return []catalogRootEntry{
		{kind: "route", identity: catalogCutoverRoute, document: route},
		{kind: "ack", identity: catalogCutoverRoute + ":frontend-v3", document: ack},
	}
}

func unretireCatalogCutoverConsumer(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	execCatalogCutoverSeed(t, ctx, database, `UPDATE catalog_a.catalog_legacy_route_hits
SET retired_at=NULL WHERE route_key=$1`, catalogCutoverRoute)
}

func blockCatalogCutoverInventory(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	execCatalogCutoverSeed(t, ctx, database, `INSERT INTO catalog_a.module_dependencies(
id,module_id,depends_on_module_id) VALUES(
'018f2000-0000-7000-8000-000000000073',
'018f2000-0000-7000-8000-000000000074',
'018f2000-0000-7000-8000-000000000075')`)
}

func unblockCatalogCutoverInventory(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	execCatalogCutoverSeed(t, ctx, database, `DELETE FROM catalog_a.module_dependencies
WHERE id='018f2000-0000-7000-8000-000000000073'`)
}

func retireCatalogCutoverConsumer(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	execCatalogCutoverSeed(t, ctx, database, `UPDATE catalog_a.catalog_legacy_route_hits
SET retired_at=$2::timestamptz WHERE route_key=$1`, catalogCutoverRoute,
		catalogCutoverRetiredAt)
}

func queryCatalogCutoverText(t *testing.T, ctx context.Context,
	database *sql.DB, query string, arguments ...any,
) string {
	t.Helper()
	var result string
	if err := database.QueryRowContext(ctx, query, arguments...).Scan(&result); err != nil {
		t.Fatal(err)
	}
	return result
}
