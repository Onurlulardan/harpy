//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBillingOfferingAvailabilityAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := billingTestContext(t)
			defer cancel()
			database := prepareBillingOrigin(t, ctx, origin)
			if _, err := database.ExecContext(ctx,
				"GRANT harpy_runtime TO harpy_v2_runtime WITH INHERIT FALSE, SET TRUE",
			); err != nil {
				t.Fatal(err)
			}
			connection := billingRuntimeConnection(t, ctx, database)
			authorityID := billingGlobalAuthorityID(t, ctx, database)
			actorID := uuid.New()
			now := time.Now().UTC().Truncate(time.Microsecond)
			created := createAvailabilityOffering(
				t, ctx, connection, authorityID, actorID, now,
			)
			if count := publicOfferingCount(t, ctx, connection, created.offeringID, now); count != 0 {
				t.Fatalf("unavailable public count = %d", count)
			}
			updated := publishAvailabilityVersion(
				t, ctx, connection, authorityID, actorID, created, now.Add(time.Second),
			)
			if updated.revision != 2 || updated.version != 2 || updated.offeringID != created.offeringID {
				t.Fatalf("updated offering = %+v", updated)
			}
			assertImmutableAvailability(t, ctx, connection, created.offeringID, now)
			future := publishFutureAvailabilityVersion(
				t, ctx, connection, authorityID, actorID, updated, now,
			)
			if future.revision != 3 || future.version != 3 {
				t.Fatalf("future offering version = %+v", future)
			}
			if _, err := connection.ExecContext(ctx, "SET ROLE harpy_runtime"); err != nil {
				t.Fatal(err)
			}
			assertCanonicalOfferingHistory(
				t, ctx, connection, created.offeringID, now,
			)
		})
	}
}

func createAvailabilityOffering(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	authorityID, actorID uuid.UUID, now time.Time,
) offeringResult {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var result offeringResult
	err = transaction.QueryRowContext(ctx, `
SELECT * FROM platform.billing_create_offering_v3(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24
)`, authorityID, billingV7(t), billingV7(t), billingV7(t), actorID,
		"availability-create", billingDigest("availability-create"),
		"availability-plan", "Availability Plan", "", 9, "", "", "monthly",
		true, false, false, 0, 0, int16(1), int16(0), now.Add(-time.Hour), nil, now,
	).Scan(&result.offeringID, &result.versionID, &result.revision,
		&result.version, &result.lifecycle, &result.replay)
	commitOfferingAvailability(t, ctx, transaction, authorityID, result.versionID, now, err)
	return result
}

func publishAvailabilityVersion(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	authorityID, actorID uuid.UUID, current offeringResult, now time.Time,
) offeringResult {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var result offeringResult
	err = transaction.QueryRowContext(ctx, `
SELECT * FROM platform.billing_create_offering_version_v3(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24
)`, authorityID, billingV7(t), current.offeringID, billingV7(t), actorID,
		"availability-publish", billingDigest("availability-publish"), current.revision,
		"Availability Plan", "", 9, "", "", "monthly", true, true, false,
		0, 0, int16(1), int16(0), now.Add(-time.Hour), nil, now,
	).Scan(&result.offeringID, &result.versionID, &result.revision,
		&result.version, &result.lifecycle, &result.replay)
	commitOfferingAvailability(t, ctx, transaction, authorityID, result.versionID, now, err)
	return result
}

func commitOfferingAvailability(
	t *testing.T, ctx context.Context, transaction *sql.Tx,
	authorityID, versionID uuid.UUID, now time.Time, callErr error,
) {
	t.Helper()
	if callErr == nil {
		callErr = addOfferingPrice(ctx, transaction, authorityID, versionID, now)
	}
	if callErr != nil {
		_ = transaction.Rollback()
		t.Fatal(callErr)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func publicOfferingCount(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	offeringID uuid.UUID, now time.Time,
) int {
	t.Helper()
	var count int
	err := connection.QueryRowContext(ctx, `
SELECT count(*) FROM platform.billing_offerings AS root
JOIN LATERAL (
 SELECT candidate.* FROM platform.billing_offering_versions AS candidate
 WHERE candidate.authority_id=root.authority_id AND candidate.offering_id=root.id
   AND candidate.publication_state='published'
 ORDER BY candidate.version DESC, candidate.id ASC LIMIT 1
) AS latest ON true
WHERE root.id=$1 AND root.lifecycle='active' AND latest.is_public AND latest.available
  AND latest.effective_start <= $2
  AND (latest.effective_end IS NULL OR latest.effective_end > $2)`, offeringID, now,
	).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}

func assertImmutableAvailability(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	offeringID uuid.UUID, now time.Time,
) {
	t.Helper()
	var lifecycle string
	var revision int64
	var versions, unavailable, available int
	err := connection.QueryRowContext(ctx, `
SELECT root.lifecycle, root.revision, count(*),
       count(*) FILTER (WHERE NOT version_row.available),
       count(*) FILTER (WHERE version_row.available)
FROM platform.billing_offerings AS root
JOIN platform.billing_offering_versions AS version_row
  ON version_row.authority_id=root.authority_id AND version_row.offering_id=root.id
WHERE root.id=$1 GROUP BY root.lifecycle, root.revision`, offeringID,
	).Scan(&lifecycle, &revision, &versions, &unavailable, &available)
	if err != nil || lifecycle != "active" || revision != 2 ||
		versions != 2 || unavailable != 1 || available != 1 ||
		publicOfferingCount(t, ctx, connection, offeringID, now) != 1 {
		t.Fatalf("immutable availability = %s/%d/%d/%d/%d err=%v",
			lifecycle, revision, versions, unavailable, available, err)
	}
}
