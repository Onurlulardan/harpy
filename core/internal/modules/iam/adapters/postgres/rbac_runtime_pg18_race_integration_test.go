//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

const raceIAMRBACAction = `
SELECT result_id::text FROM iam_rbac_action_create_v1(
    pg_catalog.current_schema()::name, $1::uuid, $2::bytea, $3::uuid,
    $4::text, $5::text, 'race test', false,
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001', 1, 4
)`

func assertIAMRBACMutationRace(t *testing.T, ctx context.Context,
	database *sql.DB, schema string) {
	t.Helper()
	type outcome struct{ err error }
	results := make(chan outcome, 2)
	inputs := [][]any{
		{"99400000-0000-4000-8000-000000000001", bytes.Repeat([]byte{0x81}, 32),
			"99500000-0000-4000-8000-000000000001", "race_one", "Race One"},
		{"99400000-0000-4000-8000-000000000002", bytes.Repeat([]byte{0x82}, 32),
			"99500000-0000-4000-8000-000000000002", "race_two", "Race Two"},
	}
	for _, input := range inputs {
		arguments := input
		go func() {
			_, err := runIAMAtomicText(ctx, database, schema, raceIAMRBACAction, arguments...)
			results <- outcome{err: err}
		}()
	}
	successes := 0
	for range inputs {
		if result := <-results; result.err == nil {
			successes++
		}
	}
	if successes != 1 || iamRBACClock(t, ctx, database, schema) != 5 {
		t.Fatalf("concurrent RBAC successes = %d", successes)
	}
	var created int
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT pg_catalog.count(*) FROM iam_actions
WHERE id IN ('99500000-0000-4000-8000-000000000001',
             '99500000-0000-4000-8000-000000000002')`, &created)
	if created != 1 {
		t.Fatalf("concurrent RBAC action count = %d", created)
	}
}
