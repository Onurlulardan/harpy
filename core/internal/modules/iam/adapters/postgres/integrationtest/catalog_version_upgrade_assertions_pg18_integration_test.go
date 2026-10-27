//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func assertIAMCatalogTwoVersionUpgrade(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	v1Document := iamCatalogDocument(t)
	v1Digest := sha256.Sum256(v1Document)
	v1Hex := hex.EncodeToString(v1Digest[:])
	assertIAMCatalogBackfilledHead(t, ctx, db, v1Hex)
	v1Replay := projectIAMCatalogVersion(
		t, ctx, db,
		"018f0000-0004-7000-8000-000000000013",
		"1.0.0", "018f0000-0004-7000-8000-000000000004",
		"project.iam.catalog.v1", v1Document, 2, 2,
		"iam-catalog-operational-request-v1",
		time.Date(2026, 7, 30, 10, 4, 0, 0, time.UTC),
	)
	if v1Replay.receiptID != "018f0000-0004-7000-8000-000000000003" ||
		v1Replay.packageVersion != "1.0.0" ||
		v1Replay.contentSHA256 != v1Hex ||
		v1Replay.documentSHA256 != v1Hex {
		t.Fatalf("IAM v1 replay differs: %+v", v1Replay)
	}
	v2Document := iamCatalogDocumentV2(t)
	v2Digest := sha256.Sum256(v2Document)
	v2Hex := hex.EncodeToString(v2Digest[:])
	v2 := projectIAMCatalogVersion(
		t, ctx, db, iamCatalogV2ReceiptID, "2.0.0",
		iamCatalogV2ReleaseID,
		iamCatalogV2OperationID+":"+iamCatalogComponentID+":provision",
		v2Document, 3, 3, "iam-catalog-version-request:2.0.0",
		time.Date(2026, 7, 30, 10, 5, 0, 0, time.UTC),
	)
	if v2.receiptID != iamCatalogV2ReceiptID ||
		v2.packageVersion != "2.0.0" ||
		v2.releaseID != iamCatalogV2ReleaseID ||
		v2.contentSHA256 != v2Hex ||
		v2.documentSHA256 != v2Hex || v2.state != "STAGED" {
		t.Fatalf("IAM v2 projection differs: %+v", v2)
	}
	assertIAMCatalogV1StillActive(t, ctx, db, v1Hex, v2Hex)
	applyIAMCatalogVersionAction(
		t, ctx, db,
		"018f0000-0004-7000-8000-000000000014",
		iamCatalogV2OperationID, "stage", v2Hex,
		time.Date(2026, 7, 30, 10, 6, 0, 0, time.UTC),
	)
	assertIAMCatalogV1StillActive(t, ctx, db, v1Hex, v2Hex)
	retry := projectIAMCatalogVersion(
		t, ctx, db, iamCatalogRetryReceipt, "2.0.0",
		iamCatalogV2ReleaseID,
		iamCatalogRetryOpID+":"+iamCatalogComponentID+":provision",
		v2Document, 3, 3, "iam-catalog-version-retry:2.0.0",
		time.Date(2026, 7, 30, 10, 6, 10, 0, time.UTC),
	)
	if retry.receiptID != iamCatalogRetryReceipt ||
		retry.contentSHA256 != v2Hex || retry.state != "STAGED" {
		t.Fatalf("IAM v2 retry projection differs: %+v", retry)
	}
	applyIAMCatalogVersionAction(
		t, ctx, db, "018f0000-0004-7000-8000-000000000019",
		iamCatalogRetryOpID, "stage", v2Hex,
		time.Date(2026, 7, 30, 10, 6, 20, 0, time.UTC),
	)
	applyIAMCatalogVersionAction(
		t, ctx, db, "018f0000-0004-7000-8000-00000000001a",
		iamCatalogV2OperationID, "cleanup", v2Hex,
		time.Date(2026, 7, 30, 10, 6, 30, 0, time.UTC),
	)
	assertIAMCatalogRetryCandidate(t, ctx, db, v2Hex)
	applyIAMCatalogVersionAction(
		t, ctx, db,
		"018f0000-0004-7000-8000-00000000001b",
		iamCatalogRetryOpID, "activate", v2Hex,
		time.Date(2026, 7, 30, 10, 7, 0, 0, time.UTC),
	)
	applyIAMCatalogVersionAction(
		t, ctx, db,
		"018f0000-0004-7000-8000-00000000001b",
		iamCatalogRetryOpID, "activate", v2Hex,
		time.Date(2026, 7, 30, 10, 7, 0, 0, time.UTC),
	)
	assertIAMCatalogV2Active(t, ctx, db, v2Hex, iamCatalogRetryReceipt)
	assertIAMCatalogVersionRead(
		t, ctx, db,
		"018f0000-0004-7000-8000-000000000003",
		"1.0.0", "018f0000-0004-7000-8000-000000000004", v1Hex,
	)
	assertIAMCatalogVersionRead(
		t, ctx, db, iamCatalogV2ReceiptID,
		"2.0.0", iamCatalogV2ReleaseID, v2Hex,
	)
	assertIAMCatalogVersionRead(
		t, ctx, db, iamCatalogRetryReceipt,
		"2.0.0", iamCatalogV2ReleaseID, v2Hex,
	)
	again := projectIAMCatalogVersion(
		t, ctx, db,
		"018f0000-0004-7000-8000-000000000016",
		"1.0.0", "018f0000-0004-7000-8000-000000000004",
		"project.iam.catalog.v1", v1Document, 2, 2,
		"iam-catalog-operational-request-v1",
		time.Date(2026, 7, 30, 10, 8, 0, 0, time.UTC),
	)
	if again.receiptID != v1Replay.receiptID {
		t.Fatalf("IAM v1 replay identity changed: %+v", again)
	}
	assertIAMCatalogV2HeadUnchanged(t, ctx, db, v2Hex)
}

func assertIAMCatalogBackfilledHead(
	t *testing.T, ctx context.Context, db publicationDatabase, v1Hex string,
) {
	t.Helper()
	const statement = `SELECT head.package_version,head.content_sha256,
 (SELECT count(*) FROM iam_catalog_version_evidence),
 (SELECT count(*) FROM iam_catalog_component_candidates)
FROM iam_catalog_component_heads AS head WHERE head.component_id=$1`
	var version, digest string
	var evidence, candidates int
	err := queryIAMCatalogTenant(t, ctx, db, statement, iamCatalogComponentID).
		Scan(&version, &digest, &evidence, &candidates)
	if err != nil || version != "1.0.0" || digest != v1Hex ||
		evidence != 1 || candidates != 0 {
		t.Fatalf("IAM backfilled head = %s/%s/%d/%d/%v",
			version, digest, evidence, candidates, err)
	}
}

func assertIAMCatalogV1StillActive(
	t *testing.T, ctx context.Context, db publicationDatabase,
	v1Hex, v2Hex string,
) {
	t.Helper()
	const statement = `SELECT head.package_version,head.content_sha256,
 component.content_sha256,component.lifecycle_state,
 candidate.package_version,candidate.content_sha256,
 (SELECT display_name FROM iam_resources WHERE resource_key='form.project-task'),
 (SELECT clock_value FROM iam_authorization_clock WHERE singleton),
 (SELECT count(*) FROM iam_catalog_version_operational_resources
  WHERE package_version='2.0.0')
FROM iam_catalog_component_heads AS head
JOIN iam_catalog_components AS component USING(component_id)
JOIN iam_catalog_component_candidates AS candidate USING(component_id)
WHERE head.component_id=$1`
	var headVersion, headDigest, componentDigest, state string
	var candidateVersion, candidateDigest, taskName string
	var clock int64
	var projected int
	err := queryIAMCatalogTenant(t, ctx, db, statement, iamCatalogComponentID).
		Scan(
			&headVersion, &headDigest, &componentDigest, &state,
			&candidateVersion, &candidateDigest, &taskName, &clock, &projected,
		)
	if err != nil || headVersion != "1.0.0" || headDigest != v1Hex ||
		componentDigest != v1Hex || state != "ACTIVE" ||
		candidateVersion != "2.0.0" || candidateDigest != v2Hex ||
		taskName != "Task" || clock != 1 || projected != 0 {
		t.Fatalf("IAM staged v2 disturbed v1: %s/%s/%s/%s/%s/%s/%s/%d/%d/%v",
			headVersion, headDigest, componentDigest, state,
			candidateVersion, candidateDigest, taskName, clock, projected, err)
	}
}
