//go:build integration

package integrationtest

import (
	"context"
	"testing"
)

func assertIAMCatalogRetryCandidate(
	t *testing.T, ctx context.Context, db publicationDatabase, v2Hex string,
) {
	t.Helper()
	const statement = `SELECT candidate.provision_receipt_id::text,
 candidate.staged_operation_id::text,
 (SELECT count(*) FROM iam_catalog_version_evidence
  WHERE package_version='2.0.0'),
 (SELECT count(*) FROM iam_catalog_version_provisions
  WHERE package_version='2.0.0'),
 (SELECT count(*) FROM iam_catalog_component_receipts
  WHERE action='stage' AND content_sha256=$2
    AND operation_id IN ($3::uuid,$4::uuid)),
 (SELECT count(*) FROM iam_catalog_component_receipts
  WHERE action='cleanup' AND operation_id=$3::uuid)
FROM iam_catalog_component_candidates AS candidate
WHERE candidate.component_id=$1::uuid`
	var receipt, operation string
	var evidence, provisions, stages, cleanups int
	err := queryIAMCatalogTenant(
		t, ctx, db, statement, iamCatalogComponentID, v2Hex,
		iamCatalogV2OperationID, iamCatalogRetryOpID,
	).Scan(
		&receipt, &operation, &evidence, &provisions, &stages, &cleanups,
	)
	if err != nil || receipt != iamCatalogRetryReceipt ||
		operation != iamCatalogRetryOpID || evidence != 1 ||
		provisions != 2 || stages != 2 || cleanups != 1 {
		t.Fatalf(
			"IAM retry candidate=%s/%s/%d/%d/%d/%d/%v",
			receipt, operation, evidence, provisions, stages, cleanups, err,
		)
	}
}
