//go:build integration

package integrationtest

import (
	"context"
	"testing"
)

func assertIAMCatalogV2Active(
	t *testing.T, ctx context.Context, db publicationDatabase,
	v2Hex, wantReceipt string,
) {
	t.Helper()
	const statement = `SELECT
 (SELECT count(*) FROM iam_catalog_version_evidence),
 (SELECT count(*) FROM iam_catalog_version_provisions
  WHERE package_version='2.0.0'),
 head.package_version,head.release_id::text,head.content_sha256,
 head.provision_receipt_id::text,
 component.content_sha256,component.lifecycle_state,
 (SELECT count(*) FROM iam_catalog_component_candidates),
 (SELECT display_name FROM iam_resources WHERE resource_key='form.project-task'),
 (SELECT clock_value FROM iam_authorization_clock WHERE singleton),
 (SELECT count(*) FROM iam_catalog_version_operational_resources
  WHERE package_version='2.0.0'),
 (SELECT count(*) FROM iam_catalog_version_operational_grants
  WHERE package_version='2.0.0')
FROM iam_catalog_component_heads AS head
JOIN iam_catalog_components AS component USING(component_id)
WHERE head.component_id=$1`
	var evidence, provisions, candidates, resources, grants int
	var version, releaseID, headDigest, headReceipt string
	var componentDigest, state, taskName string
	var clock int64
	err := queryIAMCatalogTenant(t, ctx, db, statement, iamCatalogComponentID).
		Scan(
			&evidence, &provisions, &version, &releaseID, &headDigest, &headReceipt,
			&componentDigest, &state, &candidates, &taskName,
			&clock, &resources, &grants,
		)
	if err != nil || evidence != 2 || provisions != 2 || version != "2.0.0" ||
		releaseID != iamCatalogV2ReleaseID ||
		headDigest != v2Hex || headReceipt != wantReceipt ||
		componentDigest != v2Hex ||
		state != "ACTIVE" || candidates != 0 || taskName != "Task v2" ||
		clock != 2 || resources != 3 || grants != 5 {
		t.Fatalf(
			"IAM v2 active = %d/%d/%s/%s/%s/%s/%s/%s/%d/%s/%d/%d/%d/%v",
			evidence, provisions, version, releaseID, headDigest, headReceipt,
			componentDigest,
			state, candidates, taskName, clock, resources, grants, err,
		)
	}
}

func assertIAMCatalogVersionRead(
	t *testing.T, ctx context.Context, db publicationDatabase,
	receiptID, wantVersion, wantReleaseID, wantDigest string,
) {
	t.Helper()
	const statement = `SELECT package_version,release_id::text,
 content_sha256,pg_catalog.encode(pg_catalog.sha256(document_bytes),'hex')
FROM iam_catalog_projection_read_v2($1::uuid)`
	var version, releaseID, contentDigest, documentDigest string
	err := queryIAMCatalogTenant(t, ctx, db, statement, receiptID).
		Scan(&version, &releaseID, &contentDigest, &documentDigest)
	if err != nil || version != wantVersion || releaseID != wantReleaseID ||
		contentDigest != wantDigest || documentDigest != wantDigest {
		t.Fatalf("IAM catalog exact read = %s/%s/%s/%s/%v",
			version, releaseID, contentDigest, documentDigest, err)
	}
}

func assertIAMCatalogV2HeadUnchanged(
	t *testing.T, ctx context.Context, db publicationDatabase, v2Hex string,
) {
	t.Helper()
	const statement = `SELECT head.package_version,head.content_sha256,
 (SELECT count(*) FROM iam_catalog_component_candidates)
FROM iam_catalog_component_heads AS head WHERE head.component_id=$1`
	var version, digest string
	var candidates int
	err := queryIAMCatalogTenant(t, ctx, db, statement, iamCatalogComponentID).
		Scan(&version, &digest, &candidates)
	if err != nil || version != "2.0.0" ||
		digest != v2Hex || candidates != 0 {
		t.Fatalf("IAM v1 replay moved v2 head = %s/%s/%d/%v",
			version, digest, candidates, err)
	}
}
