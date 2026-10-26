//go:build integration

package postgres

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func testStorageProviderLocator(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	now := time.Date(2026, 7, 18, 14, 0, 0, 0, time.UTC)
	account, external := bytes.Repeat([]byte{0x44}, 32), bytes.Repeat([]byte{0x55}, 32)
	tenant := "019b0000-0000-7000-8000-000000000020"
	delivery := "019b0000-0000-7000-8000-000000000021"
	var gotTenant, gotDelivery string
	for _, acceptedAt := range []time.Time{now, now.Add(time.Hour)} {
		err := connection.QueryRow(ctx, insertProviderLocatorSQL, "smtp", account, external,
			tenant, delivery, acceptedAt).Scan(&gotTenant, &gotDelivery)
		if err != nil || gotTenant != tenant || gotDelivery != delivery {
			t.Fatalf("locator replay tenant=%s delivery=%s err=%v", gotTenant, gotDelivery, err)
		}
	}
	err := connection.QueryRow(ctx, insertProviderLocatorSQL, "smtp", account, external,
		"019b0000-0000-7000-8000-000000000022",
		"019b0000-0000-7000-8000-000000000023", now).Scan(&gotTenant, &gotDelivery)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("conflicting locator did not fail closed: %v", err)
	}
}

func testStorageProviderReceipt(t *testing.T, ctx context.Context, connection *pgx.Conn) {
	t.Helper()
	now := time.Date(2026, 7, 18, 15, 0, 0, 0, time.UTC)
	message := "019b0000-0000-7000-8000-000000000040"
	delivery := "019b0000-0000-7000-8000-000000000041"
	tx := beginStorageTenant(t, ctx, connection)
	seedStorageDelivery(t, ctx, tx, now, message, delivery,
		"019b0000-0000-7000-8000-000000000042")
	if _, err := tx.Exec(ctx, claimDeliverySQL, delivery, "worker-b", now); err != nil {
		t.Fatal(err)
	}
	result, err := tx.Exec(ctx, updateDeliverySQL, delivery, int64(2), "accepted", 1,
		"", int64(1), nil, now, "ACCEPTED", false, int64(3))
	if err != nil || result.RowsAffected() != 1 {
		t.Fatalf("accept affected=%d err=%v", result.RowsAffected(), err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	reference := bytes.Repeat([]byte{0x66}, 32)
	target := bytes.Repeat([]byte{0x77}, 32)
	receipt := "019b0000-0000-7000-8000-000000000043"
	for attempt := 0; attempt < 2; attempt++ {
		tx = beginStorageTenant(t, ctx, connection)
		var got string
		var replay bool
		err = tx.QueryRow(ctx, recordProviderReceiptSQL, receipt, delivery, reference,
			"bounced", "provider-key-2026-07", now.Add(time.Second), target).
			Scan(&got, &replay)
		if err != nil || got != delivery || replay != (attempt == 1) {
			t.Fatalf("receipt attempt=%d id=%s replay=%t err=%v", attempt, got, replay, err)
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	assertProviderReceiptState(t, ctx, connection, delivery, reference, target)
}

func assertProviderReceiptState(t *testing.T, ctx context.Context, connection *pgx.Conn,
	delivery string, reference, target []byte,
) {
	t.Helper()
	tx := beginStorageTenant(t, ctx, connection)
	var state, reason string
	var revision, receipts int64
	err := tx.QueryRow(ctx, `SELECT d.state,d.row_version,
 (SELECT count(*) FROM notifications_delivery_receipts r WHERE r.delivery_id=d.id),
 (SELECT s.reason FROM notifications_suppressed_targets s
  WHERE s.channel=d.channel AND s.target_digest=$2)
 FROM notifications_deliveries d WHERE d.id=$1`, delivery, target).
		Scan(&state, &revision, &receipts, &reason)
	if err != nil || state != "bounced" || revision != 4 || receipts != 1 || reason != "bounce" {
		t.Fatalf("provider state=%s revision=%d receipts=%d reason=%s err=%v",
			state, revision, receipts, reason, err)
	}
	var ignored string
	var replay bool
	err = tx.QueryRow(ctx, recordProviderReceiptSQL,
		"019b0000-0000-7000-8000-000000000044", delivery, reference,
		"complained", "provider-key-2026-07", time.Now().UTC(), target).
		Scan(&ignored, &replay)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("conflicting receipt outcome=%v", err)
	}
	_ = tx.Rollback(ctx)
}
