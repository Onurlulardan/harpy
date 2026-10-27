//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func publishFutureAvailabilityVersion(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	authorityID, actorID uuid.UUID,
	current offeringResult,
	now time.Time,
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
		"availability-future", billingDigest("availability-future"), current.revision,
		"Availability Plan Future", "", 9, "", "", "monthly", true, true, false,
		0, 0, int16(1), int16(0), now.Add(24*time.Hour), nil, now.Add(2*time.Second),
	).Scan(&result.offeringID, &result.versionID, &result.revision,
		&result.version, &result.lifecycle, &result.replay)
	commitOfferingAvailability(
		t, ctx, transaction, authorityID, result.versionID, now.Add(2*time.Second), err,
	)
	return result
}

func assertCanonicalOfferingHistory(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	offeringID uuid.UUID,
	snapshotAt time.Time,
) {
	t.Helper()
	var rows, availableRows int
	var maximumVersion int64
	err := connection.QueryRowContext(ctx, `
SELECT count(*), count(*) FILTER (WHERE available), max(version)
FROM platform.billing_offering_query_v2
WHERE offering_id=$1`, offeringID,
	).Scan(&rows, &availableRows, &maximumVersion)
	if err != nil || rows != 3 || availableRows != 2 || maximumVersion != 3 {
		t.Fatalf("canonical history rows/available/max = %d/%d/%d err=%v",
			rows, availableRows, maximumVersion, err)
	}
	var effectiveVersions []int64
	rowsResult, err := connection.QueryContext(ctx, `
SELECT version FROM platform.billing_offering_query_v2
WHERE offering_id=$1 AND lifecycle='active' AND publication_state='published'
  AND is_public AND available AND effective_start <= $2
  AND (effective_end IS NULL OR effective_end > $2)
ORDER BY version`, offeringID, snapshotAt)
	if err != nil {
		t.Fatal(err)
	}
	defer rowsResult.Close()
	for rowsResult.Next() {
		var version int64
		if err := rowsResult.Scan(&version); err != nil {
			t.Fatal(err)
		}
		effectiveVersions = append(effectiveVersions, version)
	}
	if err := rowsResult.Err(); err != nil {
		t.Fatal(err)
	}
	if len(effectiveVersions) != 1 || effectiveVersions[0] != 2 {
		t.Fatalf("effective public versions = %v", effectiveVersions)
	}
}
