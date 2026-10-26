//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type providerPGFinalization struct {
	decision, state       string
	activeID, finalizerID uuid.UUID
	revision              int64
}

func TestProviderFinalizerClaimRotationAndRollbackOnPG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := prepareProviderPG18(t, ctx)
	fixture := prepareProviderPGFixture(t, ctx, database)
	initialAttempt := createProviderPGCommand(t, ctx, &fixture)
	assertProviderPGWildcardClaimRejected(t, ctx, fixture, initialAttempt)
	assertProviderPGMissingAttemptRejected(t, ctx, fixture, initialAttempt)
	assertProviderPGTerminalZeroAttemptRejected(t, ctx, fixture, initialAttempt)

	nextOne, reconcileOne := billingPGV7(t), billingPGV7(t)
	transaction, err := fixture.connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	first := finalizeProviderPG(t, ctx, transaction, fixture,
		fixture.finalizerID, initialAttempt, nextOne, reconcileOne)
	if first.decision != "pending" || first.finalizerID != nextOne ||
		first.revision != 2 {
		t.Fatalf("first finalization = %#v", first)
	}
	nextOneAttempt := insertProviderPGFinalizer(t, ctx, transaction, fixture, nextOne)
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}

	assertProviderPGStaleFinalizer(t, ctx, fixture, initialAttempt)
	nextTwo, reconcileTwo := billingPGV7(t), billingPGV7(t)
	transaction, err = fixture.connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	second := finalizeProviderPG(t, ctx, transaction, fixture,
		nextOne, nextOneAttempt, nextTwo, reconcileTwo)
	if second.decision != "pending" || second.finalizerID != nextTwo ||
		second.revision != 3 {
		t.Fatalf("successor finalization = %#v", second)
	}
	nextTwoAttempt := insertProviderPGFinalizer(t, ctx, transaction, fixture, nextTwo)
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}

	nextThree, reconcileThree := billingPGV7(t), billingPGV7(t)
	transaction, err = fixture.connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	rolledBack := finalizeProviderPG(t, ctx, transaction, fixture,
		nextTwo, nextTwoAttempt, nextThree, reconcileThree)
	if rolledBack.finalizerID != nextThree {
		t.Fatalf("rollback candidate = %#v", rolledBack)
	}
	if err = transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	var storedFinalizer uuid.UUID
	var revision int64
	if err = fixture.connection.QueryRowContext(ctx, `
SELECT finalizer_job_id,revision FROM platform.billing_provider_commands WHERE id=$1
`, fixture.commandID).Scan(&storedFinalizer, &revision); err != nil {
		t.Fatal(err)
	}
	if storedFinalizer != nextTwo || revision != 3 {
		t.Fatalf("rolled back ledger finalizer=%s revision=%d", storedFinalizer, revision)
	}
}

func finalizeProviderPG(
	t *testing.T,
	ctx context.Context,
	runner providerPGRunner,
	f providerPGFixture,
	jobID, attemptID, nextID, reconciliationID uuid.UUID,
) providerPGFinalization {
	t.Helper()
	var result providerPGFinalization
	err := runner.QueryRowContext(ctx, `
SELECT decision,active_job_id,finalizer_job_id,state,revision
FROM platform.billing_finalize_provider_command_v1(
 $1,$2,$3,1,1,$4,$5,$6,$7,$8,$9
)`, f.tenantID, jobID, attemptID,
		"billing-provider-finalize/"+f.commandID.String(), f.digestHex,
		f.commandID, nextID, reconciliationID, f.now.Add(10*time.Second),
	).Scan(&result.decision, &result.activeID, &result.finalizerID,
		&result.state, &result.revision)
	if err != nil {
		t.Fatalf("finalize provider command: %v", err)
	}
	return result
}

func assertProviderPGStaleFinalizer(
	t *testing.T,
	ctx context.Context,
	f providerPGFixture,
	attemptID uuid.UUID,
) {
	t.Helper()
	err := f.connection.QueryRowContext(ctx, `
SELECT decision FROM platform.billing_finalize_provider_command_v1(
 $1,$2,$3,1,1,$4,$5,$6,$7,$8,$9
)`, f.tenantID, f.finalizerID, attemptID,
		"billing-provider-finalize/"+f.commandID.String(), f.digestHex,
		f.commandID, billingPGV7(t), billingPGV7(t), f.now.Add(11*time.Second),
	).Scan(new(string))
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) || pgError.Code != "P5731" {
		t.Fatalf("stale finalizer error=%v", err)
	}
}

var _ providerPGRunner = (*sql.Tx)(nil)
