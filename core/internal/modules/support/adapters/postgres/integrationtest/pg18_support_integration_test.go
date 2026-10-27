//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestSupportFreshAndAdoptedCommandsOnPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startSupportDatabase(t, ctx)
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			resetAndApplySupport(t, ctx, database, origin)
			exerciseContactLifecycle(t, ctx, database)
			exerciseBugIsolation(t, ctx, database)
			assertRuntimePrivileges(t, ctx, database)
		})
	}
}

func exerciseContactLifecycle(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	now := time.Date(2026, 8, 1, 10, 11, 12, 123456000, time.UTC)
	id := uuid.Must(uuid.NewV7())
	rate := sha256.Sum256([]byte("support-rate-fixture"))
	digest := sha256.Sum256([]byte("support-contact-fixture"))
	var document []byte
	var replayed, admitted bool
	query := `SELECT response_document,replayed,admitted FROM platform.support_submit_contact_v1(
        $1,'Ada','ada@example.com','','','','Need support','general','integration',$2,$3,$4,$5)`
	if err := database.QueryRowContext(ctx, query, id, rate[:], "contact-command-0001",
		digest[:], now).Scan(&document, &replayed, &admitted); err != nil {
		t.Fatal(err)
	}
	if replayed || !admitted || len(document) == 0 {
		t.Fatalf("unexpected contact result replay=%v admitted=%v", replayed, admitted)
	}
	var replayDocument []byte
	if err := database.QueryRowContext(ctx, query, uuid.Must(uuid.NewV7()), rate[:],
		"contact-command-0001", digest[:], now).Scan(&replayDocument, &replayed, &admitted); err != nil {
		t.Fatal(err)
	}
	if !replayed || !admitted || string(replayDocument) != string(document) {
		t.Fatal("contact replay did not return the exact stored response")
	}
	wrong := sha256.Sum256([]byte("different-contact"))
	err := database.QueryRowContext(ctx, query, uuid.Must(uuid.NewV7()), rate[:],
		"contact-command-0001", wrong[:], now).Scan(&replayDocument, &replayed, &admitted)
	assertSQLState(t, err, "P6502")
	for index := 2; index <= 6; index++ {
		key := "contact-command-000" + string(rune('0'+index))
		requestDigest := sha256.Sum256([]byte(key))
		err = database.QueryRowContext(ctx, query, uuid.Must(uuid.NewV7()), rate[:], key,
			requestDigest[:], now.Add(time.Duration(index)*time.Microsecond)).
			Scan(&replayDocument, &replayed, &admitted)
		if err != nil || admitted != (index <= 5) || replayed {
			t.Fatalf("rate request %d admitted=%v replay=%v err=%v", index, admitted, replayed, err)
		}
	}
	var used int
	if err = database.QueryRowContext(ctx, `SELECT request_count FROM platform.support_public_rate_leases
        WHERE rate_key_sha256=$1`, rate[:]).Scan(&used); err != nil || used != 6 {
		t.Fatalf("persisted rate count=%d err=%v, want 6", used, err)
	}
}

func exerciseBugIsolation(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	now := time.Date(2026, 8, 1, 11, 12, 13, 654321000, time.UTC)
	tenantA, tenantB, actor := uuid.New(), uuid.New(), uuid.New()
	first := createBug(t, ctx, database, tenantA, actor, "tenant-a", "Tenant A", now, "bug-command-00001")
	_ = createBug(t, ctx, database, tenantB, actor, "tenant-b", "Tenant B", now.Add(time.Microsecond), "bug-command-00002")
	var items []byte
	var total int64
	if err := database.QueryRowContext(ctx, `SELECT items,total FROM platform.support_list_bugs_v1(
        $1,1,1,false,$2,NULL,NULL,NULL,NULL,1,20)`, actor, tenantA).Scan(&items, &total); err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("tenant A total = %d, want 1", total)
	}
	var hidden []byte
	err := database.QueryRowContext(ctx, `SELECT response_document FROM platform.support_get_bug_v1(
        $1,1,1,$2,$3,$4)`, actor, tenantA, uuid.New(), first).Scan(&hidden)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cross-reporter read error = %v, want no rows", err)
	}
	digest := sha256.Sum256([]byte("bug-update"))
	var updated []byte
	var replayed bool
	update := `SELECT response_document,replayed FROM platform.support_update_bug_v1(
        $1,1,1,false,$2,$3,true,'resolved',false,NULL,false,NULL,true,'fixed',1,$4,$5,$6)`
	if err := database.QueryRowContext(ctx, update, actor, tenantA, first,
		"bug-update-command-01", digest[:], now.Add(2*time.Microsecond)).Scan(&updated, &replayed); err != nil {
		t.Fatal(err)
	}
	if replayed || len(updated) == 0 {
		t.Fatal("bug update was not committed")
	}
	err = database.QueryRowContext(ctx, update, actor, tenantA, first,
		"bug-update-command-02", digest[:], now.Add(3*time.Microsecond)).Scan(&updated, &replayed)
	assertSQLState(t, err, "P6503")
}

func createBug(t *testing.T, ctx context.Context, database *sql.DB, tenant, actor uuid.UUID,
	slug, name string, observed time.Time, key string,
) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	digest := sha256.Sum256([]byte(key))
	var document []byte
	var replayed bool
	err := database.QueryRowContext(ctx, `SELECT response_document,replayed FROM platform.support_create_bug_v1(
        $1,$2,$3,$4,$5,'Ada','ada@example.com',1,1,'Broken page','The page is broken','bug','high',
        'https://example.com/page','browser','{}','[]','[]','Open the page',$6,$7,$8)`,
		id, tenant, slug, name, actor, key, digest[:], observed).Scan(&document, &replayed)
	if err != nil || replayed || len(document) == 0 {
		t.Fatalf("create bug: replay=%v document=%d err=%v", replayed, len(document), err)
	}
	return id
}

func assertRuntimePrivileges(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err = connection.ExecContext(ctx, `SET ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	_, err = connection.ExecContext(ctx, `SELECT id FROM platform.support_bug_reports LIMIT 1`)
	assertSQLState(t, err, "42501")
	_, err = connection.ExecContext(ctx, `SELECT platform.support_claim_adoption_v1(
        '00000000-0000-0000-0000-000000000001','00000000-0000-7000-8000-000000000001',
        '00000000-0000-0000-0000-000000000002',1,pg_catalog.now())`)
	assertSQLState(t, err, "42501")
	var stats []byte
	if err = connection.QueryRowContext(ctx, `SELECT platform.support_bug_stats_v1(
        '20000000-0000-4000-8000-000000000001',1,1,true,NULL)`).Scan(&stats); err != nil || len(stats) == 0 {
		t.Fatalf("runtime routine execution failed: %v", err)
	}
	if _, resetErr := connection.ExecContext(ctx, `RESET ROLE`); resetErr != nil {
		t.Fatal(resetErr)
	}
}

func assertSQLState(t *testing.T, err error, expected string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != expected {
		t.Fatalf("SQLSTATE = %v, want %s", err, expected)
	}
}
