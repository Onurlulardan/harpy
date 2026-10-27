package postgres

import (
	"strings"
	"testing"
)

func TestHTTPIdempotencySQLUsesTransactionLockDBClockAndAppendOnlyInsert(t *testing.T) {
	t.Parallel()
	for _, marker := range []string{
		"pg_catalog.pg_try_advisory_xact_lock($1::bigint)",
		"FROM platform.idempotency_receipts", "tenant_id = $2", "principal_id = $3",
		"operation_id = $4", "idempotency_key = $5", "record_kind = 'http'",
	} {
		if !strings.Contains(readHTTPIdempotencyReceiptSQL, marker) {
			t.Fatalf("read SQL lacks %q", marker)
		}
	}
	for _, marker := range []string{
		"pg_catalog.clock_timestamp()", "interval '24 hours'", "interval '30 days'",
		"GREATEST", "INSERT INTO platform.idempotency_receipts",
	} {
		if !strings.Contains(insertHTTPIdempotencyReceiptSQL, marker) {
			t.Fatalf("insert SQL lacks %q", marker)
		}
	}
	upper := strings.ToUpper(insertHTTPIdempotencyReceiptSQL)
	if strings.Contains(upper, "UPDATE ") || strings.Contains(upper, "DELETE ") {
		t.Fatal("HTTP receipt persistence is not append-only")
	}
}
