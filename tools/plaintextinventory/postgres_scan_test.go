package main

import (
	"strings"
	"testing"
)

func TestDatabaseColumnQueryQuotesEveryIdentifier(t *testing.T) {
	query := databaseColumnQuery(databaseColumn{
		schema: `tenant";DROP SCHEMA platform;--`,
		table:  `events table`, column: `payload"value`, dataType: "jsonb",
	})
	for _, required := range []string{
		`"tenant"";DROP SCHEMA platform;--"."events table"`,
		`"payload""value"`,
		`position($1::bytea IN pg_catalog.convert_to(`,
	} {
		if !strings.Contains(query, required) {
			t.Fatalf("query is missing %q: %s", required, query)
		}
	}
}

func TestDatabaseByteaColumnDoesNotCastThroughText(t *testing.T) {
	query := databaseColumnQuery(databaseColumn{
		schema: "platform", table: "receipts", column: "payload", dataType: "bytea",
	})
	if strings.Contains(query, "convert_to") ||
		!strings.Contains(query, `position($1::bytea IN "payload")`) {
		t.Fatalf("bytea query = %s", query)
	}
}
