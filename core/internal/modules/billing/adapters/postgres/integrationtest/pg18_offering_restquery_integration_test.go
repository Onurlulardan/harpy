//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBillingOfferingRESTQueryViewAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := billingTestContext(t)
	defer cancel()
	database := startBillingPG18(t, ctx, "restquery")
	prepareBillingPrerequisites(t, ctx, database, "fresh")
	applyBillingDocuments(t, ctx, database, billingRESTQueryDocuments(t, "fresh"))
	if _, err := database.ExecContext(ctx,
		"GRANT harpy_runtime TO harpy_v2_runtime WITH INHERIT FALSE, SET TRUE",
	); err != nil {
		t.Fatal(err)
	}
	connection := billingRuntimeConnection(t, ctx, database)
	authorityID := billingGlobalAuthorityID(t, ctx, database)
	actorID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	public := seedRESTQueryOffering(
		t, ctx, connection, authorityID, actorID, "rest-public", true, now,
	)
	seedRESTQueryOffering(t, ctx, connection, authorityID, actorID, "rest-private", false, now)
	discount := createDiscount(
		t, ctx, connection, authorityID, actorID, public.versionID, now,
	)
	campaign := createCampaign(
		t, ctx, connection, authorityID, actorID, discount.versionID, "rest-campaign", now,
	)

	assertOfferingQueryCount(t, ctx, connection, "", 0)
	assertCatalogQueryCount(t, ctx, connection, "billing_discount_query_v1", 0)
	assertCatalogQueryCount(t, ctx, connection, "billing_campaign_query_v1", 0)
	if _, err := connection.ExecContext(ctx, "SET ROLE harpy_runtime"); err != nil {
		t.Fatal(err)
	}
	assertOfferingQueryCount(t, ctx, connection, "WHERE offering_key LIKE 'rest-%'", 2)
	assertOfferingQueryCount(t, ctx, connection, `WHERE offering_key LIKE 'rest-%'
AND lifecycle='active' AND publication_state='published' AND is_public
AND effective_start <= pg_catalog.now()
AND (effective_end IS NULL OR effective_end > pg_catalog.now())`, 1)
	assertCatalogQueryCount(t, ctx, connection, "billing_discount_query_v1", 1)
	assertCatalogQueryCount(t, ctx, connection, "billing_campaign_query_v1", 1)
	transaction := beginBillingRuntimeSnapshot(t, ctx, connection)
	assertCatalogSnapshotReads(
		t, ctx, transaction, public, discount, campaign, now,
	)
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}

	if _, err := database.ExecContext(ctx, `UPDATE platform.cutover_groups
SET generation=2 WHERE scope_kind='global' AND group_id='billing'`); err != nil {
		t.Fatal(err)
	}
	assertOfferingQueryCount(t, ctx, connection, "", 0)
	assertCatalogQueryCount(t, ctx, connection, "billing_discount_query_v1", 0)
	assertCatalogQueryCount(t, ctx, connection, "billing_campaign_query_v1", 0)
}

func seedRESTQueryOffering(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	authorityID, actorID uuid.UUID, key string, public bool, now time.Time,
) offeringResult {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var result offeringResult
	err = transaction.QueryRowContext(ctx, `
SELECT * FROM platform.billing_create_offering_v1(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18
)`, authorityID, billingV7(t), billingV7(t), billingV7(t), actorID,
		key, billingDigest(key), key, "REST Query "+key, public, false,
		0, 0, int16(1), int16(0), now.Add(-time.Hour), nil, now,
	).Scan(&result.offeringID, &result.versionID, &result.revision,
		&result.version, &result.lifecycle, &result.replay)
	if err == nil {
		err = addOfferingPrice(ctx, transaction, authorityID, result.versionID, now)
	}
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return result
}

func assertOfferingQueryCount(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	where string, expected int,
) {
	t.Helper()
	query := "SELECT count(*) FROM platform.billing_offering_query_v1 " + where
	var count int
	if err := connection.QueryRowContext(ctx, query).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != expected {
		t.Fatalf("offering query count = %d, want %d", count, expected)
	}
}

func assertCatalogQueryCount(
	t *testing.T, ctx context.Context, connection *sql.Conn, view string, expected int,
) {
	t.Helper()
	var query string
	switch view {
	case "billing_discount_query_v1":
		query = "SELECT count(*) FROM platform.billing_discount_query_v1"
	case "billing_campaign_query_v1":
		query = "SELECT count(*) FROM platform.billing_campaign_query_v1"
	default:
		t.Fatalf("unsupported Billing catalog query view %q", view)
	}
	var count int
	if err := connection.QueryRowContext(ctx, query).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != expected {
		t.Fatalf("%s count = %d, want %d", view, count, expected)
	}
}

func billingRESTQueryDocuments(t *testing.T, origin string) []billingDocument {
	t.Helper()
	if origin != "fresh" && origin != "v1_adopted" {
		t.Fatalf("Billing migration origin %q is unsupported", origin)
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(workingDirectory, "..", "..", "..", "migrations", "platform", "fresh")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	documents := make([]billingDocument, 0, len(entries))
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".sql") {
			path := filepath.Join(directory, entry.Name())
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			documents = append(documents, billingDocument{path: path, content: content})
		}
	}
	sort.Slice(documents, func(left, right int) bool { return documents[left].path < documents[right].path })
	return documents
}
