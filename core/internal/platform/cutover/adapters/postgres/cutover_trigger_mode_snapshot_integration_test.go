//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

type userTriggerModeSnapshot struct {
	tableName, triggerName, mode string
}

func captureUserTriggerModes(
	t *testing.T, ctx context.Context, tx pgx.Tx, schema string, tables []string,
) []userTriggerModeSnapshot {
	t.Helper()
	rows, err := tx.Query(ctx, `SELECT relation.relname,trigger_row.tgname,
trigger_row.tgenabled::text
FROM pg_catalog.pg_trigger AS trigger_row
JOIN pg_catalog.pg_class AS relation ON relation.oid=trigger_row.tgrelid
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
WHERE namespace.nspname=$1 AND relation.relname=ANY($2::text[])
  AND NOT trigger_row.tgisinternal
ORDER BY relation.relname,trigger_row.tgname`, schema, tables)
	if err != nil {
		t.Fatalf("capture user trigger modes: %v", err)
	}
	defer rows.Close()
	var snapshots []userTriggerModeSnapshot
	for rows.Next() {
		var snapshot userTriggerModeSnapshot
		if err := rows.Scan(
			&snapshot.tableName, &snapshot.triggerName, &snapshot.mode,
		); err != nil {
			t.Fatalf("scan user trigger mode: %v", err)
		}
		snapshots = append(snapshots, snapshot)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read user trigger modes: %v", err)
	}
	return snapshots
}

func restoreUserTriggerModes(
	t *testing.T, ctx context.Context, tx pgx.Tx, schema string,
	snapshots []userTriggerModeSnapshot,
) {
	t.Helper()
	actions := map[string]string{
		"D": "DISABLE", "O": "ENABLE", "R": "ENABLE REPLICA", "A": "ENABLE ALWAYS",
	}
	for _, snapshot := range snapshots {
		action, ok := actions[snapshot.mode]
		if !ok {
			t.Fatalf("unsupported trigger mode %q for %s", snapshot.mode, snapshot.triggerName)
		}
		statement := `ALTER TABLE ` +
			pgx.Identifier{schema, snapshot.tableName}.Sanitize() + ` ` + action +
			` TRIGGER ` + pgx.Identifier{snapshot.triggerName}.Sanitize()
		if _, err := tx.Exec(ctx, statement); err != nil {
			t.Fatalf("restore trigger %s.%s mode %s: %v",
				snapshot.tableName, snapshot.triggerName, snapshot.mode, err)
		}
	}
}
