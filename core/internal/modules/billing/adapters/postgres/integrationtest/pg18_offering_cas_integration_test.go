//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

func assertOfferingCAS(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	authorityID uuid.UUID,
	actorID uuid.UUID,
	created offeringResult,
	now time.Time,
) {
	t.Helper()
	connections := []*sql.Conn{
		billingRuntimeConnection(t, ctx, database),
		billingRuntimeConnection(t, ctx, database),
	}
	start := make(chan struct{})
	results := make(chan offeringCall, 2)
	for index, connection := range connections {
		go func(index int, connection *sql.Conn) {
			<-start
			results <- callPublishOffering(ctx, connection, authorityID, actorID,
				created.offeringID, index, now)
		}(index, connection)
	}
	close(start)
	first, second := <-results, <-results
	successes, stale := 0, 0
	for _, call := range []offeringCall{first, second} {
		switch {
		case call.err == nil && call.result.revision == 2:
			successes++
		case billingPGCode(call.err) == "40001":
			stale++
		default:
			t.Fatalf("unexpected concurrent publication: result=%+v err=%v",
				call.result, call.err)
		}
	}
	if successes != 1 || stale != 1 {
		t.Fatalf("publication race successes=%d stale=%d", successes, stale)
	}
}

func callPublishOffering(
	ctx context.Context,
	connection *sql.Conn,
	authorityID uuid.UUID,
	actorID uuid.UUID,
	offeringID uuid.UUID,
	index int,
	now time.Time,
) offeringCall {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return offeringCall{err: err}
	}
	result := offeringResult{}
	err = transaction.QueryRowContext(ctx, `
SELECT * FROM platform.billing_create_offering_version_v1(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18
)`, authorityID, billingV7Must(), offeringID, billingV7Must(), actorID,
		"offering-publish-"+string(rune('a'+index)),
		billingDigest("offering-publish-"+string(rune('a'+index))), int64(1),
		"Integration Plan V2", true, false, 0, 0, int16(1), int16(0), now, nil, now,
	).Scan(&result.offeringID, &result.versionID, &result.revision,
		&result.version, &result.lifecycle, &result.replay)
	if err != nil {
		_ = transaction.Rollback()
		return offeringCall{result: result, err: err}
	}
	if err = addOfferingPrice(ctx, transaction, authorityID, result.versionID, now); err != nil {
		_ = transaction.Rollback()
		return offeringCall{result: result, err: err}
	}
	err = transaction.Commit()
	return offeringCall{result: result, err: err}
}
