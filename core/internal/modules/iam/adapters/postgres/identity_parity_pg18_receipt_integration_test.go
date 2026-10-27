//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

const iamBulkDeleteReserveSQL = `SELECT iam_identity_bulk_delete_reserve_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,'user',$3,2)`

func assertIAMIdentityBulkDeleteReceipt(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	operation := "77000000-0000-4000-8000-000000000001"
	digest := bytes.Repeat([]byte{0x77}, 32)
	withIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(tx *sql.Tx) {
		first, err := reserveIAMBulkDelete(
			ctx, tx, operation, iamIdentityActorID, digest,
		)
		if err != nil || first {
			t.Fatalf("first bulk receipt=%t error=%v", first, err)
		}
		replay, err := reserveIAMBulkDelete(
			ctx, tx, operation, iamIdentityActorID, digest,
		)
		if err != nil || !replay {
			t.Fatalf("bulk receipt replay=%t error=%v", replay, err)
		}
	})
	probeIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(tx *sql.Tx) {
		_, err := reserveIAMBulkDelete(
			ctx, tx, operation, iamIdentityActorID,
			bytes.Repeat([]byte{0x78}, 32),
		)
		assertIAMSQLState(t, err, "23505")
	})
	var count int
	mustQueryIAMTenantValue(t, ctx, database, "t_iam_a", `
SELECT count(*) FROM iam_identity_bulk_delete_receipts
WHERE operation_id='77000000-0000-4000-8000-000000000001'`, &count)
	if count != 1 {
		t.Fatalf("durable bulk delete receipt count=%d", count)
	}
	assertIAMIdentityBulkDeleteReceiptRace(t, ctx, database)
}

func assertIAMIdentityBulkDeleteReceiptRace(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	operation := "77000000-0000-4000-8000-000000000002"
	digest := bytes.Repeat([]byte{0x79}, 32)
	connection, transaction, err := beginIAMParityRuntimeTx(
		ctx, database, "t_iam_b",
	)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	first, err := reserveIAMBulkDelete(
		ctx, transaction, operation, iamIdentityActorID, digest,
	)
	if err != nil || first {
		_ = transaction.Rollback()
		t.Fatalf("race first receipt=%t error=%v", first, err)
	}
	result := make(chan iamBulkDeleteRaceResult, 1)
	go reserveIAMBulkDeleteConcurrent(
		ctx, database, operation, digest, result,
	)
	select {
	case early := <-result:
		_ = transaction.Rollback()
		t.Fatalf("concurrent receipt did not wait: %+v", early)
	case <-time.After(250 * time.Millisecond):
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	if _, err := connection.ExecContext(ctx, "RESET SESSION AUTHORIZATION"); err != nil {
		t.Fatal(err)
	}
	select {
	case concurrent := <-result:
		if concurrent.err != nil || !concurrent.replayed {
			t.Fatalf("concurrent receipt=%+v", concurrent)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent receipt remained blocked after commit")
	}
}

type iamBulkDeleteRaceResult struct {
	replayed bool
	err      error
}

func reserveIAMBulkDeleteConcurrent(
	ctx context.Context, database *sql.DB, operation string, digest []byte,
	result chan<- iamBulkDeleteRaceResult,
) {
	connection, transaction, err := beginIAMParityRuntimeTx(
		ctx, database, "t_iam_b",
	)
	if err != nil {
		result <- iamBulkDeleteRaceResult{err: err}
		return
	}
	defer connection.Close()
	replayed, err := reserveIAMBulkDelete(
		ctx, transaction, operation, iamIdentityActorID, digest,
	)
	if err == nil {
		err = transaction.Commit()
	} else {
		_ = transaction.Rollback()
	}
	if _, resetErr := connection.ExecContext(
		ctx, "RESET SESSION AUTHORIZATION",
	); err == nil {
		err = resetErr
	}
	result <- iamBulkDeleteRaceResult{replayed: replayed, err: err}
}

func beginIAMParityRuntimeTx(
	ctx context.Context, database *sql.DB, schema string,
) (*sql.Conn, *sql.Tx, error) {
	connection, err := database.Conn(ctx)
	if err != nil {
		return nil, nil, err
	}
	if _, err = connection.ExecContext(
		ctx, "SET SESSION AUTHORIZATION harpy_v2_runtime",
	); err != nil {
		connection.Close()
		return nil, nil, err
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err == nil {
		_, err = transaction.ExecContext(ctx,
			"SELECT pg_catalog.set_config('search_path',$1,true)",
			"pg_catalog, "+schema)
	}
	if err != nil {
		connection.Close()
		return nil, nil, err
	}
	return connection, transaction, nil
}

func reserveIAMBulkDelete(
	ctx context.Context, transaction *sql.Tx,
	operation, actor string, digest []byte,
) (bool, error) {
	var replayed bool
	err := transaction.QueryRowContext(
		ctx, iamBulkDeleteReserveSQL, operation, actor, digest,
	).Scan(&replayed)
	return replayed, err
}
