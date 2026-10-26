//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"path"
	"strings"
	"testing"
	"time"

	platformmigration "github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

type analyticsLegacyBlockerProof struct {
	receiptID, operationID, componentID  string
	catalogVersion, zeroDigest, blockers string
	referenceEpoch                       int64
	completedAt                          time.Time
}

func splitAnalyticsPurgeBridgeDocuments(
	t *testing.T, origin string,
) ([]platformmigration.Document, []platformmigration.Document) {
	t.Helper()
	documents := analyticsDocuments(t, origin, "tenant")
	for index, document := range documents {
		if strings.HasPrefix(path.Base(document.Path()), "20260730143045_") {
			return documents[:index], documents[index:]
		}
	}
	t.Fatal("Analytics purge receipt bridge migration is missing")
	return nil, nil
}

func insertAnalyticsLegacyBlockerProof(
	t *testing.T, ctx context.Context, database *sql.DB,
	receiptID, operationID, componentID, blockers string, completedAt time.Time,
) analyticsLegacyBlockerProof {
	t.Helper()
	_, err := database.ExecContext(ctx, `
WITH proof AS (SELECT $5::jsonb AS blockers)
INSERT INTO t_analytics_a.analytics_catalog_blocker_receipts(
 receipt_id,operation_id,component_id,reference_epoch,catalog_version,
 zero_digest,blockers,completed_at
)
SELECT $1,$2,$3,1,'analytics-catalog-blockers.v1',
 encode(sha256(convert_to(
  'analytics-catalog-blockers.v1'||E'\\000'||$3::uuid::text||E'\\000'||
  '1'||E'\\000'||proof.blockers::text,'UTF8'
 )),'hex'),proof.blockers,$4
FROM proof`, receiptID, operationID, componentID, completedAt, blockers)
	if err != nil {
		t.Fatalf("insert Analytics legacy blocker proof: %v", err)
	}
	return readAnalyticsLegacyBlockerProof(
		t, ctx, database, operationID, componentID,
	)
}

func readAnalyticsLegacyBlockerProof(
	t *testing.T, ctx context.Context, database *sql.DB,
	operationID, componentID string,
) analyticsLegacyBlockerProof {
	t.Helper()
	var proof analyticsLegacyBlockerProof
	err := database.QueryRowContext(ctx, `SELECT
 receipt_id::text,operation_id::text,component_id::text,reference_epoch,
 catalog_version,zero_digest,blockers::text,completed_at
FROM t_analytics_a.analytics_catalog_blocker_receipts
WHERE operation_id=$1 AND component_id=$2`,
		operationID, componentID,
	).Scan(&proof.receiptID, &proof.operationID, &proof.componentID,
		&proof.referenceEpoch, &proof.catalogVersion, &proof.zeroDigest,
		&proof.blockers, &proof.completedAt)
	if err != nil {
		t.Fatalf("read Analytics legacy blocker proof: %v", err)
	}
	return proof
}

func requireAnalyticsLegacyBlockerProofUnchanged(
	t *testing.T, ctx context.Context, database *sql.DB,
	want analyticsLegacyBlockerProof,
) {
	t.Helper()
	got := readAnalyticsLegacyBlockerProof(
		t, ctx, database, want.operationID, want.componentID,
	)
	if got.receiptID != want.receiptID || got.operationID != want.operationID ||
		got.componentID != want.componentID ||
		got.referenceEpoch != want.referenceEpoch ||
		got.catalogVersion != want.catalogVersion ||
		got.zeroDigest != want.zeroDigest || got.blockers != want.blockers ||
		!got.completedAt.Equal(want.completedAt) {
		t.Fatalf("Analytics legacy blocker proof changed: got=%+v want=%+v", got, want)
	}
}

func assertAnalyticsLegacyZeroEpochRollback(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	fixture analyticsCatalogPurgeFixture, operationID, receiptID string,
	completedAt time.Time,
) {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		t.Fatal(err)
	}
	if err = queryAnalyticsCatalogCommandTx(
		ctx, transaction, fixture.endpoint, receiptID,
		operationID, "purge", 0, completedAt,
	); err != nil {
		t.Fatalf("Analytics legacy zero-epoch compatibility: %v", err)
	}
}
