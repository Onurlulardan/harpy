//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func runConcurrentFormLinkRegisters(t *testing.T, ctx context.Context,
	database *sql.DB, schema, statement string, now time.Time,
) {
	t.Helper()
	start := make(chan struct{})
	results := make(chan error, 2)
	for index := 0; index < 2; index++ {
		index := index
		go func() {
			tx, err := database.BeginTx(ctx, nil)
			if err == nil {
				_, err = tx.ExecContext(ctx,
					"SELECT pg_catalog.set_config('search_path',$1,true)",
					schema+", pg_catalog")
			}
			<-start
			if err == nil {
				id := 4 + index*3
				_, err = tx.ExecContext(ctx, statement,
					formLinkUUID(id), formLinkActor, formLinkItem, "FORMS",
					"form_record", formLinkSource, formLinkRecord, "LIFECYCLE",
					formLinkUUID(id+1), int64(9), now,
					formLinkHash(schema, index), "form-link-race-"+string(rune('a'+index)),
					formLinkUUID(id+2))
			}
			if err == nil {
				err = tx.Commit()
			} else if tx != nil {
				_ = tx.Rollback()
			}
			results <- err
		}()
	}
	close(start)
	for index := 0; index < 2; index++ {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := database.QueryRowContext(ctx, `SELECT count(*) FROM `+schema+
		`.documents_links WHERE item_id=$1 AND lifecycle_state='ACTIVE'`,
		formLinkItem).Scan(&count); err != nil || count != 1 {
		t.Fatalf("%s concurrent active links=%d err=%v", schema, count, err)
	}
}

func assertFormLinkRollback(t *testing.T, ctx context.Context, database *sql.DB,
	schema, statement string, now time.Time,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `CREATE TABLE `+schema+
		`.forms_record_atomic_probe (id uuid PRIMARY KEY)`)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.ExecContext(ctx, "SELECT pg_catalog.set_config('search_path',$1,true)",
		schema+", pg_catalog")
	if err == nil {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO forms_record_atomic_probe (id) VALUES ($1)`, formLinkRecord)
	}
	var values [12]any
	if err == nil {
		row := tx.QueryRowContext(ctx, statement,
			formLinkUUID(30), formLinkActor, formLinkUUID(31), "FORMS",
			"form_record", formLinkSource, formLinkRecord, "LIFECYCLE",
			formLinkUUID(32), int64(1), now, formLinkHash(schema, 9),
			"form-link-rollback", formLinkUUID(33))
		destinations := make([]any, len(values))
		for index := range values {
			destinations[index] = &values[index]
		}
		err = row.Scan(destinations...)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		t.Fatalf("%s missing-item link error=%v", schema, err)
	}
	if err = tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = database.QueryRowContext(ctx, `SELECT count(*) FROM `+schema+
		`.forms_record_atomic_probe`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("%s rolled-back record count=%d err=%v", schema, count, err)
	}
}

func registerFormLinkSQL(t *testing.T, ctx context.Context, tx *sql.Tx,
	statement, linkID, operationID, auditID, key string, now time.Time,
) {
	t.Helper()
	hash := sha256.Sum256([]byte(key))
	if _, err := tx.ExecContext(ctx, statement, linkID, formLinkActor, formLinkItem,
		"FORMS", "form_record", formLinkSource, formLinkRecord, "LIFECYCLE",
		operationID, int64(9), now, hash[:], key, auditID); err != nil {
		t.Fatal(err)
	}
}

func readDocumentsLinkSQL(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "..",
		filepath.FromSlash(documentsLinkSQLRoot+name))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func formLinkHash(schema string, index int) []byte {
	value := sha256.Sum256([]byte(schema + ":" + string(rune('0'+index))))
	return value[:]
}

func formLinkUUID(value int) string {
	return "018f0000-0000-7000-8000-0000000009" + []string{
		"00", "01", "02", "03", "04", "05", "06", "07", "08", "09",
		"10", "11", "12", "13", "14", "15", "16", "17", "18", "19",
		"20", "21", "22", "23", "24", "25", "26", "27", "28", "29",
		"30", "31", "32", "33",
	}[value]
}
