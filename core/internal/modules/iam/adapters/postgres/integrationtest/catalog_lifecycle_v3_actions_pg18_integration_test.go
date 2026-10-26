//go:build integration

package integrationtest

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

const (
	iamCatalogDisableOperation = "018f0000-0004-7000-8000-000000000021"
	iamCatalogEnableOperation  = "018f0000-0004-7000-8000-000000000023"
	iamCatalogRetireOperation  = "018f0000-0004-7000-8000-000000000025"
	iamCatalogPurgeOperation   = "018f0000-0004-7000-8000-000000000027"
	iamCatalogExternalBinding  = "018f0000-0004-7000-8000-000000000029"
)

func iamCatalogLifecycleV3Documents(
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
			return version >= 20260730145300 && version <= 20260730145307
		},
	)
}

func callIAMCatalogLifecycleAction(
	ctx context.Context, db publicationDatabase,
	receiptID, operationID, action, contentSHA256 string,
	referenceEpoch int64, completedAt time.Time,
) (string, bool, error) {
	tx, err := db.admin.Begin(ctx)
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(ctx, "SET LOCAL ROLE harpy_runtime"); err != nil {
		return "", false, err
	}
	if _, err = tx.Exec(
		ctx, "SELECT pg_catalog.set_config('search_path',$1,true)",
		"pg_catalog,"+db.tenant.schema+",platform",
	); err != nil {
		return "", false, err
	}
	const statement = `SELECT applied.receipt_id::text,applied.ready
FROM iam_catalog_component_apply_v1(
 $1::uuid,$2::uuid,$3::uuid,'iam',1,'iam.catalog-template.v1',
 $4,$5,$6,$7,$8
) AS applied`
	var storedReceipt string
	var ready bool
	err = tx.QueryRow(
		ctx, statement, receiptID, operationID, iamCatalogComponentID,
		contentSHA256, action, referenceEpoch,
		operationID+":"+iamCatalogComponentID+":"+action, completedAt,
	).Scan(&storedReceipt, &ready)
	if err != nil {
		return "", false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return "", false, err
	}
	return storedReceipt, ready, nil
}

func callIAMCatalogBlockersV2(
	ctx context.Context, db publicationDatabase,
	receiptID, operationID string, referenceEpoch int64,
	completedAt time.Time,
) (string, json.RawMessage, error) {
	tx, err := db.admin.Begin(ctx)
	if err != nil {
		return "", nil, err
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(ctx, "SET LOCAL ROLE harpy_runtime"); err != nil {
		return "", nil, err
	}
	if _, err = tx.Exec(
		ctx, "SELECT pg_catalog.set_config('search_path',$1,true)",
		"pg_catalog,"+db.tenant.schema+",platform",
	); err != nil {
		return "", nil, err
	}
	const statement = `SELECT blocker.receipt_id::text,blocker.blockers
FROM iam_catalog_component_blockers_v1(
 $1::uuid,$2::uuid,$3::uuid,$4,1,$5
) AS blocker`
	var storedReceipt string
	var blockers json.RawMessage
	err = tx.QueryRow(
		ctx, statement, receiptID, operationID, iamCatalogComponentID,
		referenceEpoch, completedAt,
	).Scan(&storedReceipt, &blockers)
	if err != nil {
		return "", nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return "", nil, err
	}
	return storedReceipt, blockers, nil
}
