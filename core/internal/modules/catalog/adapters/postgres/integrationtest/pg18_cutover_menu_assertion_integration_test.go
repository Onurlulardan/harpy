//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertCatalogCutoverMenuPreserved(t *testing.T, ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	var revision, version, sortOrder int
	var definitionID, versionID, itemID, content, label string
	err := database.QueryRowContext(ctx, `SELECT definition.definition_id,
definition.revision,version.version_id,version.version_number,
version.content_sha256,item.item_id,item.label,item.sort_order
FROM catalog_a.catalog_menu_definitions AS definition
JOIN catalog_a.catalog_menu_versions AS version
  ON version.version_id=definition.active_version_id
 AND version.definition_id=definition.definition_id
JOIN catalog_a.catalog_menu_items AS item
  ON item.version_id=version.version_id
WHERE definition.definition_id=$1`, catalogMenuDefinitionID).Scan(
		&definitionID, &revision, &versionID, &version, &content,
		&itemID, &label, &sortOrder,
	)
	if err != nil || definitionID != catalogMenuDefinitionID || revision != 1 ||
		versionID != catalogMenuVersionID || version != 1 ||
		content != catalogManifestDigest || itemID != catalogMenuItemID ||
		label != "Legacy operations" || sortOrder != 10 {
		t.Fatalf("Catalog adopted menu=%s/%d %s/%d/%s %s/%s/%d err=%v",
			definitionID, revision, versionID, version, content, itemID,
			label, sortOrder, err)
	}
}
