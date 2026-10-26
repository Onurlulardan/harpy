//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	iamactorv1 "github.com/Onurlulardan/harpy/compat/iamactorv1"
	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const actorBridgeBusinessRowID = "78000000-0000-4000-8000-000000000001"

type actorBridgeCompatTransaction struct{ transaction pgx.Tx }

func (adapter *actorBridgeCompatTransaction) QueryRow(
	ctx context.Context, statement string, arguments ...any,
) iamactorv1.Row {
	return adapter.transaction.QueryRow(ctx, statement, arguments...)
}

func assertActorBridgeDirectDMLClosed(
	t *testing.T, ctx context.Context, v1 *pgx.Conn,
) {
	t.Helper()
	_, err := v1.Exec(ctx,
		"DELETE FROM platform.iam_v1_actor_assertion_receipts WHERE false")
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
		t.Fatalf("direct actor receipt DML error = %v", err)
	}
}

func assertActorBridgeRollbackCommitReplay(
	t *testing.T,
	ctx context.Context,
	database cutoverIntegrationDatabase,
	v1 *pgx.Conn,
	issued actorBridgeIssuedCredential,
) {
	t.Helper()
	consumer, err := iamactorv1.NewConsumer(issued.publicKeys, func() time.Time {
		return issued.currentTime
	})
	if err != nil {
		t.Fatal(err)
	}
	defer consumer.Close()
	transaction, err := v1.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	consumeActorBridgeInTransaction(t, ctx, consumer, transaction, issued,
		actorBridgeBusinessRowID)
	if err := transaction.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	assertActorBridgeAtomicCounts(t, ctx, database.admin, 0, 0)

	transaction, err = v1.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	consumeActorBridgeInTransaction(t, ctx, consumer, transaction, issued,
		actorBridgeBusinessRowID)
	if err := transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	assertActorBridgeAtomicCounts(t, ctx, database.admin, 1, 1)

	transaction, err = v1.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	adapter := &actorBridgeCompatTransaction{transaction: transaction}
	_, err = consumer.Consume(ctx, adapter, issued.wire, issued.expectation,
		func(context.Context, iamactorv1.Transaction, actorassertionv1.Claims) error {
			return errors.New("replay reached business effect")
		})
	if !iamactorv1.Is(err, iamactorv1.CodeReplay) {
		t.Fatalf("committed assertion replay error = %v", err)
	}
	if err := transaction.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
}

func consumeActorBridgeInTransaction(
	t *testing.T,
	ctx context.Context,
	consumer *iamactorv1.Consumer,
	transaction pgx.Tx,
	issued actorBridgeIssuedCredential,
	rowID string,
) {
	t.Helper()
	adapter := &actorBridgeCompatTransaction{transaction: transaction}
	receipt, err := consumer.Consume(ctx, adapter, issued.wire, issued.expectation,
		actorBridgeBusinessEffect(rowID))
	if err != nil || receipt.ConsumedAt().IsZero() {
		t.Fatalf("consume actor assertion: receipt=%v err=%v", receipt, err)
	}
}

func actorBridgeBusinessEffect(rowID string) iamactorv1.Effect {
	return func(
		ctx context.Context,
		transaction iamactorv1.Transaction,
		_ actorassertionv1.Claims,
	) error {
		adapter, ok := transaction.(*actorBridgeCompatTransaction)
		if !ok {
			return errors.New("unexpected actor bridge transaction adapter")
		}
		_, err := adapter.transaction.Exec(ctx, `
INSERT INTO t_actor_a.actor_bridge_business_rows (id, created_at)
VALUES ($1,clock_timestamp())`, rowID)
		return err
	}
}

func assertActorBridgeAtomicCounts(
	t *testing.T, ctx context.Context, admin *pgx.Conn, wantReceipts, wantRows int64,
) {
	t.Helper()
	var receipts, rows int64
	if err := admin.QueryRow(ctx,
		"SELECT count(*) FROM platform.iam_v1_actor_assertion_receipts",
	).Scan(&receipts); err != nil {
		t.Fatal(err)
	}
	if err := admin.QueryRow(ctx,
		"SELECT count(*) FROM t_actor_a.actor_bridge_business_rows",
	).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if receipts != wantReceipts || rows != wantRows {
		t.Fatalf("actor bridge atomic counts = %d/%d, want %d/%d",
			receipts, rows, wantReceipts, wantRows)
	}
}
