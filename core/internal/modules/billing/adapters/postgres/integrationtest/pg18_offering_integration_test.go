//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
)

type offeringResult struct {
	offeringID uuid.UUID
	versionID  uuid.UUID
	revision   int64
	version    int64
	lifecycle  string
	replay     bool
}

type billingQueryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func TestBillingOfferingRoutinesAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := billingTestContext(t)
	defer cancel()
	database := prepareBillingOrigin(t, ctx, "fresh")
	connection := billingRuntimeConnection(t, ctx, database)
	authorityID := billingGlobalAuthorityID(t, ctx, database)
	actorID := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	created := createOffering(t, ctx, connection, authorityID, actorID,
		"offering-create-1", billingDigest("offering-create"), now)
	if created.revision != 1 || created.version != 1 || created.replay {
		t.Fatalf("unexpected create result: %+v", created)
	}
	replayed := callCreateOffering(ctx, connection, authorityID, actorID,
		"offering-create-1", billingDigest("offering-create"), now)
	if replayed.err != nil || !replayed.result.replay ||
		replayed.result.offeringID != created.offeringID {
		t.Fatalf("unexpected replay: result=%+v err=%v", replayed.result, replayed.err)
	}
	mismatch := callCreateOffering(ctx, connection, authorityID, actorID,
		"offering-create-1", billingDigest("different-request"), now)
	if billingPGCode(mismatch.err) != "23505" {
		t.Fatalf("idempotency mismatch code = %q, err=%v", billingPGCode(mismatch.err), mismatch.err)
	}
	assertOfferingCAS(t, ctx, database, authorityID, actorID, created, now)
	assertTenantAuthorityCannotWriteGlobal(t, ctx, connection, authorityID, actorID, now)
}

type offeringCall struct {
	result offeringResult
	err    error
}

func createOffering(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	authorityID uuid.UUID, actorID uuid.UUID, key string, digest []byte, now time.Time,
) offeringResult {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	call := callCreateOffering(ctx, transaction, authorityID, actorID, key, digest, now)
	if call.err != nil {
		_ = transaction.Rollback()
		t.Fatal(call.err)
	}
	if err = addOfferingPrice(ctx, transaction, authorityID, call.result.versionID, now); err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return call.result
}

func callCreateOffering(
	ctx context.Context, connection billingQueryRower, authorityID uuid.UUID,
	actorID uuid.UUID, key string, digest []byte, now time.Time,
) offeringCall {
	result := offeringResult{}
	err := connection.QueryRowContext(ctx, `
SELECT * FROM platform.billing_create_offering_v1(
 $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18
)`, authorityID, billingV7Must(), billingV7Must(), billingV7Must(), actorID,
		key, digest, "integration-plan", "Integration Plan", true, false,
		0, 0, int16(1), int16(0), now, nil, now,
	).Scan(&result.offeringID, &result.versionID, &result.revision,
		&result.version, &result.lifecycle, &result.replay)
	return offeringCall{result: result, err: err}
}

func addOfferingPrice(
	ctx context.Context,
	transaction *sql.Tx,
	authorityID uuid.UUID,
	versionID uuid.UUID,
	now time.Time,
) error {
	_, err := transaction.ExecContext(ctx, `
SELECT platform.billing_add_offering_price_v1($1,$2,$3,$4,$5,$6,$7,$8)
`, authorityID, versionID, billingV7Must(), "monthly", int64(1000), "USD", int64(1), now)
	return err
}

func billingV7Must() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
