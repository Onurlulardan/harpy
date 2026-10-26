//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestNotificationsPG18StorageReliability(t *testing.T) {
	ctx := context.Background()
	connection := startStoragePG18(t, ctx)
	prepareStorageDatabase(t, ctx, connection)
	t.Run("ticket is single use and exact authority bound", func(t *testing.T) {
		testStorageTicket(t, ctx, connection)
	})
	t.Run("stale delivery becomes outcome unknown atomically", func(t *testing.T) {
		testStaleStorageDelivery(t, ctx, connection)
	})
	t.Run("provider locator conflicts fail closed", func(t *testing.T) {
		testStorageProviderLocator(t, ctx, connection)
	})
	t.Run("authenticated provider receipt is idempotent", func(t *testing.T) {
		testStorageProviderReceipt(t, ctx, connection)
	})
	t.Run("message and delivery rollback together", func(t *testing.T) {
		testStorageRollback(t, ctx, connection)
	})
	t.Run("retention preserves dedupe and redacts ticket audit", func(t *testing.T) {
		testStorageRetention(t, ctx, connection)
	})
}

func testStorageTicket(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	token := "ticket-secret-not-returned"
	hash := sha256.Sum256([]byte(token))
	tenant := "019b0000-0000-7000-8000-000000000010"
	user := "019b0000-0000-7000-8000-000000000011"
	_, err := connection.Exec(ctx, insertRealtimeTicketSQL,
		"019b0000-0000-7000-8000-000000000012", tenant, user, hash[:], int64(7),
		"https://app.example.com", now, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if consumeStorageTicket(ctx, connection, hash[:], tenant, user, 8,
		"https://app.example.com", now.Add(time.Second)) {
		t.Fatal("wrong authorization version consumed ticket")
	}
	if consumeStorageTicket(ctx, connection, hash[:], tenant, user, 7,
		"https://other.example.com", now.Add(2*time.Second)) {
		t.Fatal("wrong origin consumed ticket")
	}
	if !consumeStorageTicket(ctx, connection, hash[:], tenant, user, 7,
		"https://app.example.com", now.Add(3*time.Second)) {
		t.Fatal("exact authority failed to consume ticket")
	}
	if consumeStorageTicket(ctx, connection, hash[:], tenant, user, 7,
		"https://app.example.com", now.Add(4*time.Second)) {
		t.Fatal("ticket consumed twice")
	}
}

func consumeStorageTicket(ctx context.Context, connection *pgx.Conn, hash []byte,
	tenant, user string, version int64, origin string, at time.Time,
) bool {
	var values [9]any
	err := connection.QueryRow(ctx, consumeRealtimeTicketSQL, hash, tenant, user,
		version, origin, at).Scan(&values[0], &values[1], &values[2], &values[3],
		&values[4], &values[5], &values[6], &values[7], &values[8])
	if errors.Is(err, pgx.ErrNoRows) {
		return false
	}
	return err == nil
}

func testStaleStorageDelivery(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	now := time.Date(2026, 7, 18, 13, 0, 0, 0, time.UTC)
	tx := beginStorageTenant(t, ctx, connection)
	seedStorageDelivery(t, ctx, tx, now, storageMessageID, storageDeliveryID,
		"019b0000-0000-7000-8000-000000000003")
	if _, err := tx.Exec(ctx, claimDeliverySQL, storageDeliveryID, "worker-a", now); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	tx = beginStorageTenant(t, ctx, connection)
	var reconciled string
	if err := tx.QueryRow(ctx, reconcileStaleDeliverySQL, storageDeliveryID,
		now.Add(2*time.Minute)).Scan(&reconciled); err != nil {
		t.Fatal(err)
	}
	result, err := tx.Exec(ctx, updateDeliverySQL, storageDeliveryID, int64(2), "accepted",
		1, "", int64(1), nil, now, "ACCEPTED", false, int64(3))
	if err != nil || result.RowsAffected() != 0 {
		t.Fatalf("stale finalizer affected=%d err=%v", result.RowsAffected(), err)
	}
	var state, code string
	var unknown bool
	var rowVersion, attempts int64
	err = tx.QueryRow(ctx, `SELECT d.state,d.outcome_code,d.outcome_unknown,d.row_version,
 (SELECT count(*) FROM notifications_delivery_attempts a WHERE a.delivery_id=d.id)
 FROM notifications_deliveries d WHERE d.id=$1`, storageDeliveryID).
		Scan(&state, &code, &unknown, &rowVersion, &attempts)
	if err != nil || reconciled != storageDeliveryID || state != "failed" ||
		code != "NOTIFICATION_OUTCOME_UNKNOWN" || !unknown || rowVersion != 3 || attempts != 1 {
		t.Fatalf("reconcile id=%s state=%s code=%s unknown=%t version=%d attempts=%d err=%v",
			reconciled, state, code, unknown, rowVersion, attempts, err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	assertAttemptImmutable(t, ctx, connection)
}

func beginStorageTenant(t *testing.T, ctx context.Context, connection *pgx.Conn) pgx.Tx {
	t.Helper()
	tx, err := connection.Begin(ctx)
	if err == nil {
		_, err = tx.Exec(ctx, "SELECT pg_catalog.set_config('search_path',$1,true)",
			storageTenantSchema+", pg_catalog")
	}
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

func assertAttemptImmutable(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	tx := beginStorageTenant(t, ctx, connection)
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE harpy_runtime"); err != nil {
		t.Fatal(err)
	}
	_, err := tx.Exec(ctx, `UPDATE notifications_delivery_attempts
 SET response_class='rewritten' WHERE delivery_id=$1`, storageDeliveryID)
	if err == nil {
		t.Fatal("runtime mutated immutable delivery attempt")
	}
	_ = tx.Rollback(ctx)
}

func testStorageRollback(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	tx := beginStorageTenant(t, ctx, connection)
	message := "019b0000-0000-7000-8000-000000000030"
	delivery := "019b0000-0000-7000-8000-000000000031"
	seedStorageDelivery(t, ctx, tx, time.Now().UTC(), message, delivery,
		"019b0000-0000-7000-8000-000000000032")
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	tx = beginStorageTenant(t, ctx, connection)
	var count int64
	if err := tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM notifications_messages WHERE id=$1)
 + (SELECT count(*) FROM notifications_deliveries WHERE id=$2)`, message, delivery).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rolled back rows=%d", count)
	}
	_ = tx.Rollback(ctx)
}
