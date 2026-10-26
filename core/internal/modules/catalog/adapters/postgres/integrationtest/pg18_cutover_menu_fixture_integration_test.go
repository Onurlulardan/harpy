//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

const catalogLegacyMenuID = "018f2000-0000-7000-8000-000000000080"
const catalogMenuDefinitionID = "018f2000-0000-7000-8000-000000000081"
const catalogMenuVersionID = "018f2000-0000-7000-8000-000000000082"
const catalogMenuItemID = "018f2000-0000-7000-8000-000000000083"

func loadCatalogCutoverMenuMappings(t *testing.T, ctx context.Context,
	database *sql.DB,
) []catalogCutoverMapping {
	t.Helper()
	execCatalogCutoverSeed(t, ctx, database, `INSERT INTO catalog_a.menu_configs(
id,tenant_id,config,deleted_at) VALUES($1,$2,
'{"items":[{"label":"Legacy operations","kind":"GROUP","sort_order":10}]}'::jsonb,
NULL)`, catalogLegacyMenuID, catalogTenantA)
	source := queryCatalogCutoverText(t, ctx, database, `SELECT to_jsonb(row)::text
FROM catalog_a.menu_configs AS row WHERE row.id=$1`, catalogLegacyMenuID)
	definition := queryCatalogCutoverText(t, ctx, database, `SELECT
pg_catalog.jsonb_build_object(
'definition_id',$1::uuid,'revision',1,'active_version_id',$2::uuid,
'updated_at','2026-07-21T10:00:00Z'::timestamptz)::text`,
		catalogMenuDefinitionID, catalogMenuVersionID)
	version := queryCatalogCutoverText(t, ctx, database, `SELECT
pg_catalog.jsonb_build_object(
'version_id',$1::uuid,'definition_id',$2::uuid,'version_number',1,
'content_sha256',$3::text,
'created_at','2026-07-21T10:00:00Z'::timestamptz)::text`,
		catalogMenuVersionID, catalogMenuDefinitionID, catalogManifestDigest)
	item := queryCatalogCutoverText(t, ctx, database, `SELECT
pg_catalog.jsonb_build_object(
'version_id',$1::uuid,'item_id',$2::uuid,'parent_item_id',NULL,
'package_id',$3::uuid,'item_kind','GROUP','target_id',NULL,
'external_url',NULL,'label','Legacy operations','sort_order',10)::text`,
		catalogMenuVersionID, catalogMenuItemID, catalogPackageID)
	return []catalogCutoverMapping{
		{checkpoint: "menu_configs", sourceIdentity: catalogLegacyMenuID,
			sourceDocument: source, targetKind: "menu_definition",
			targetIdentity: catalogMenuDefinitionID, targetRecord: definition},
		{checkpoint: "menu_configs", sourceIdentity: catalogLegacyMenuID,
			sourceDocument: source, targetKind: "menu_version",
			targetIdentity: catalogMenuVersionID, targetRecord: version},
		{checkpoint: "menu_configs", sourceIdentity: catalogLegacyMenuID,
			sourceDocument: source, targetKind: "menu_item",
			targetIdentity: catalogMenuVersionID + ":" + catalogMenuItemID,
			targetRecord:   item},
	}
}
