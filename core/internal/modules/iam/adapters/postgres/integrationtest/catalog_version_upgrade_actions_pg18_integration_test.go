//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/jackc/pgx/v5"
)

type iamCatalogProjection struct {
	receiptID      string
	packageVersion string
	releaseID      string
	contentSHA256  string
	documentSHA256 string
	state          string
	readyAt        time.Time
}

func iamCatalogVersionUpgradeDocuments(
	t *testing.T, origin string,
) []migration.Document {
	t.Helper()
	documents, err := iammodule.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	return selectPublicationDocuments(
		t, documents, origin, "tenant",
		func(version int64) bool {
			return version >= 20260730142011 && version <= 20260730142022
		},
	)
}

func activateLegacyIAMCatalog(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	digest := sha256.Sum256(iamCatalogDocument(t))
	const statement = `SELECT applied.action,applied.ready
FROM iam_catalog_component_apply_v1(
 $1::uuid,$2::uuid,$3::uuid,$4,$5,$6,$7,$8,$9,$10,$11
) AS applied`
	steps := []struct {
		receipt string
		action  string
		at      time.Time
	}{
		{
			"018f0000-0004-7000-8000-000000000007", "stage",
			time.Date(2026, 7, 30, 10, 2, 0, 0, time.UTC),
		},
		{
			"018f0000-0004-7000-8000-000000000008", "activate",
			time.Date(2026, 7, 30, 10, 3, 0, 0, time.UTC),
		},
	}
	for _, step := range steps {
		var action string
		var ready bool
		err := queryIAMCatalogTenant(t, ctx, db, statement,
			step.receipt, iamCatalogV1OperationID, iamCatalogComponentID,
			"iam", int64(1), catalogpackagev1.ContractIAM,
			hex.EncodeToString(digest[:]), step.action, int64(0),
			iamCatalogV1OperationID+":"+iamCatalogComponentID+":"+step.action,
			step.at,
		).Scan(&action, &ready)
		if err != nil || action != step.action || !ready {
			t.Fatalf("legacy IAM catalog %s = %s/%t/%v",
				step.action, action, ready, err)
		}
	}
}

func projectIAMCatalogVersion(
	t *testing.T, ctx context.Context, db publicationDatabase,
	proposedReceipt, packageVersion, releaseID, idempotencyKey string,
	document []byte, itemCount, auxiliaryCount int,
	requestSeed string,
	occurredAt time.Time,
) iamCatalogProjection {
	t.Helper()
	contentSHA256 := sha256.Sum256(document)
	requestSHA256 := sha256.Sum256([]byte(requestSeed))
	const statement = `SELECT projection.receipt_id::text,
 projection.package_version,projection.release_id::text,
 projection.content_sha256,
 pg_catalog.encode(pg_catalog.sha256(projection.document_bytes),'hex'),
 projection.lifecycle_state,projection.ready_at
FROM iam_catalog_project_v2(
 $1::name,$2::uuid,$3::uuid,$4,$5,$6::uuid,$7::uuid,$8,$9,$10,
 $11,$12,$13,$14,$15,$16,$17,$18,$19,$20
) AS projection`
	var result iamCatalogProjection
	err := queryIAMCatalogRuntime(t, ctx, db, statement,
		db.tenant.schema, proposedReceipt,
		"e3000000-0000-4000-8000-000000000001",
		"project", packageVersion, releaseID, iamCatalogComponentID, int64(1),
		catalogpackagev1.ContractIAM, hex.EncodeToString(contentSHA256[:]),
		"IAM_TEMPLATE", "project", "project", contentSHA256[:],
		itemCount, auxiliaryCount, document, idempotencyKey,
		requestSHA256[:], occurredAt,
	).Scan(
		&result.receiptID, &result.packageVersion, &result.releaseID,
		&result.contentSHA256, &result.documentSHA256,
		&result.state, &result.readyAt,
	)
	if err != nil {
		t.Fatalf("project IAM catalog version %s: %v", packageVersion, err)
	}
	return result
}

func applyIAMCatalogVersionAction(
	t *testing.T, ctx context.Context, db publicationDatabase,
	receiptID, operationID, action, contentSHA256 string, completedAt time.Time,
) {
	t.Helper()
	const statement = `SELECT applied.action,applied.ready
FROM iam_catalog_component_apply_v2(
 $1::uuid,$2::uuid,$3::uuid,$4,$5,$6,$7,$8,$9,$10,$11
) AS applied`
	var actual string
	var ready bool
	err := queryIAMCatalogRuntime(t, ctx, db, statement,
		receiptID, operationID, iamCatalogComponentID,
		"iam", int64(1), catalogpackagev1.ContractIAM,
		contentSHA256, action, int64(0),
		operationID+":"+iamCatalogComponentID+":"+action,
		completedAt,
	).Scan(&actual, &ready)
	if err != nil || actual != action || !ready {
		t.Fatalf("IAM catalog version %s = %s/%t/%v",
			action, actual, ready, err)
	}
}

func queryIAMCatalogRuntime(
	t *testing.T, ctx context.Context, db publicationDatabase,
	statement string, arguments ...any,
) pgx.Row {
	t.Helper()
	tx, err := db.admin.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	if _, err = tx.Exec(ctx, "SET LOCAL ROLE harpy_runtime"); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(
		ctx, "SELECT pg_catalog.set_config('search_path',$1,true)",
		"pg_catalog,"+db.tenant.schema+",platform",
	); err != nil {
		t.Fatal(err)
	}
	return &commitRow{Row: tx.QueryRow(ctx, statement, arguments...), tx: tx, t: t}
}
