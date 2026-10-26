//go:build integration

package integrationtest

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBillingRetentionLegalHoldAndReplayOnPG18(t *testing.T) {
	ctx, cancel := billingTestContext(t)
	defer cancel()
	fixture := prepareRetentionFixture(t, ctx)
	if _, err := fixture.connection.ExecContext(
		ctx, "SET ROLE harpy_runtime",
	); err != nil {
		t.Fatal(err)
	}
	holdID := billingV7(t)
	var placedID uuid.UUID
	err := fixture.connection.QueryRowContext(ctx, `
SELECT hold_id FROM platform.billing_place_legal_hold_v1(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10
)`, fixture.authorityA, fixture.tenantA, holdID, billingV7(t), uuid.New(),
		"hold-2026-legal", billingDigest("place-hold"), "litigation",
		"LEGAL-2026-41", time.Now().UTC(),
	).Scan(&placedID)
	if err != nil || placedID != holdID {
		t.Fatalf("place hold=%s err=%v", placedID, err)
	}
	held := executeRetention(
		t, ctx, fixture.connection, fixture.authorityA, fixture.tenantA,
		billingV7(t), billingV7(t), billingV7(t),
		"billing-retention://held", retentionDigest("c"),
	)
	if held.held != 1 || held.redacted != 0 || held.checkpoint != 1 {
		t.Fatalf("held retention result=%+v", held)
	}
	var releasedID uuid.UUID
	err = fixture.connection.QueryRowContext(ctx, `
SELECT hold_id FROM platform.billing_release_legal_hold_v1(
 $1,$2,$3,$4,$5,$6,$7,$8,$9
)`, fixture.authorityA, fixture.tenantA, holdID, billingV7(t), uuid.New(),
		"release-2026-legal", billingDigest("release-hold"),
		"court released preservation order", time.Now().UTC(),
	).Scan(&releasedID)
	if err != nil || releasedID != holdID {
		t.Fatalf("release hold=%s err=%v", releasedID, err)
	}
	jobID, successorID := billingV7(t), billingV7(t)
	completed := executeRetention(
		t, ctx, fixture.connection, fixture.authorityA, fixture.tenantA,
		jobID, billingV7(t), successorID,
		"billing-retention://released", retentionDigest("d"),
	)
	replayed := executeRetention(
		t, ctx, fixture.connection, fixture.authorityA, fixture.tenantA,
		jobID, billingV7(t), billingV7(t),
		"billing-retention://released", retentionDigest("d"),
	)
	if completed.held != 0 || completed.redacted != 1 ||
		completed.checkpoint != 2 || replayed.successor != successorID ||
		replayed.runAt != completed.runAt ||
		replayed.redacted != completed.redacted ||
		replayed.checkpoint != completed.checkpoint {
		t.Fatalf("completed=%+v replayed=%+v", completed, replayed)
	}
	err = fixture.connection.QueryRowContext(ctx, `
SELECT checkpoint_revision
FROM platform.billing_execute_retention_cycle_v1($1,$2,$3,$4,$5,$6,$7)
`, fixture.authorityA, fixture.tenantA, jobID, billingV7(t), billingV7(t),
		"billing-retention://released", retentionDigest("e"),
	).Scan(new(int64))
	if billingPGCode(err) != "P5730" {
		t.Fatalf("reused job code=%q err=%v", billingPGCode(err), err)
	}
}
