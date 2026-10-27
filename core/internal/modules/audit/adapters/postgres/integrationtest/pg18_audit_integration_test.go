//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestAuditFreshAndAdoptedRoutinesAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startAuditDatabase(t, ctx)
	connection, _ := applyAuditOrigin(t, ctx, database, "fresh")
	exerciseAuditLifecycle(t, ctx, connection)
	exerciseEmptyAuditAdoption(t, ctx, connection)
	assertAuditRuntimeBoundary(t, ctx, connection)
}

func exerciseEmptyAuditAdoption(t *testing.T, ctx context.Context, connection *sql.Conn) {
	t.Helper()
	if _, err := connection.ExecContext(ctx, `SET ROLE harpy_adopter`); err != nil {
		t.Fatal(err)
	}
	var document []byte
	err := connection.QueryRowContext(ctx, `SELECT audit_v1_adopt_batch_v1($1,$2,10,$3)`,
		uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()),
		time.Date(2026, 8, 1, 18, 0, 0, 0, time.UTC)).Scan(&document)
	if err != nil || !containsAll(string(document), `"status": "completed"`, `"source_count": 0`) {
		t.Fatalf("empty adoption = %s, err = %v", document, err)
	}
	if _, err = connection.ExecContext(ctx, `RESET ROLE`); err != nil {
		t.Fatal(err)
	}
}

func exerciseAuditLifecycle(t *testing.T, ctx context.Context, connection *sql.Conn) {
	t.Helper()
	now := time.Date(2026, 8, 1, 16, 0, 0, 123456000, time.UTC)
	id, ownerID, actorID := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	event := sha256.Sum256([]byte("audit-event-one"))
	before, after := sha256.Sum256([]byte("old-secret")), sha256.Sum256([]byte("new-secret"))
	changes := `[{"field":"password","classification":"secret","before_sha256":"` +
		hex.EncodeToString(before[:]) + `","after_sha256":"` + hex.EncodeToString(after[:]) + `"}]`
	query := `SELECT id,replay FROM audit_entry_append_v1(
        $1,$2,$3,NULL,NULL,'iam','iam.user.update','user','user-1',NULL,'update',3,
        'profile_update',$4::jsonb,$5,$6,$6)`
	var persisted uuid.UUID
	var replay bool
	if err := connection.QueryRowContext(ctx, query, id, ownerID, actorID,
		changes, event[:], now).Scan(&persisted, &replay); err != nil || replay || persisted != id {
		t.Fatalf("first append id=%s replay=%v err=%v", persisted, replay, err)
	}
	if err := connection.QueryRowContext(ctx, query, uuid.Must(uuid.NewV7()), ownerID, actorID,
		changes, event[:], now).Scan(&persisted, &replay); err != nil || !replay || persisted != id {
		t.Fatalf("replay id=%s replay=%v err=%v", persisted, replay, err)
	}
	wrong := sha256.Sum256([]byte("different-event"))
	err := connection.QueryRowContext(ctx, query, uuid.Must(uuid.NewV7()), ownerID, actorID,
		changes, wrong[:], now).Scan(&persisted, &replay)
	assertAuditSQLState(t, err, "P6802")

	var total int64
	if err = connection.QueryRowContext(ctx, `SELECT total_count FROM audit_entries_list_v1(
        NULL,'user','user-1',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,1,20)`).Scan(&total); err != nil || total != 1 {
		t.Fatalf("list total=%d err=%v", total, err)
	}
	var stats []byte
	if err = connection.QueryRowContext(ctx, `SELECT audit_stats_v1($1)`, now.Add(time.Hour)).Scan(&stats); err != nil || len(stats) == 0 {
		t.Fatalf("stats bytes=%d err=%v", len(stats), err)
	}
	_, err = connection.ExecContext(ctx, `UPDATE audit_entries SET reason_code='tampered' WHERE id=$1`, id)
	assertAuditSQLState(t, err, "42501")
	_, err = connection.ExecContext(ctx, `DELETE FROM audit_entries WHERE id=$1`, id)
	assertAuditSQLState(t, err, "42501")
}

func assertAuditRuntimeBoundary(t *testing.T, ctx context.Context, connection *sql.Conn) {
	t.Helper()
	if _, err := connection.ExecContext(ctx, `SET ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	_, err := connection.ExecContext(ctx, `SELECT id FROM audit_entries LIMIT 1`)
	assertAuditSQLState(t, err, "42501")
	var stats []byte
	if err = connection.QueryRowContext(ctx, `SELECT audit_stats_v1($1)`,
		time.Date(2026, 8, 1, 17, 0, 0, 0, time.UTC)).Scan(&stats); err != nil || len(stats) == 0 {
		t.Fatalf("runtime stats bytes=%d err=%v", len(stats), err)
	}
	if _, err = connection.ExecContext(ctx, `RESET ROLE`); err != nil {
		t.Fatal(err)
	}
}

func assertAuditSQLState(t *testing.T, err error, expected string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != expected {
		t.Fatalf("SQLSTATE = %v, want %s", err, expected)
	}
}
