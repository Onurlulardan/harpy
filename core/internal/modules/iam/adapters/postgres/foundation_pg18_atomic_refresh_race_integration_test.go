//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"sort"
	"testing"
)

const callIAMAtomicRefresh = `SELECT iam_refresh_rotate_v1(
    $1, '93000000-0000-4000-8000-000000000001', 1,
    $2, $3, 'refresh-hmac-v1', $4
)`

func assertIAMRefreshRotationRace(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		assertIAMRefreshCrashRollback(t, ctx, database, tenant.schema)
		results := make(chan string, 2)
		errors := make(chan error, 2)
		for index := byte(1); index <= 2; index++ {
			index := index
			go func() {
				id := "93000000-0000-4000-8000-000000000010"
				if index == 2 {
					id = "93000000-0000-4000-8000-000000000011"
				}
				outcome, err := runIAMAtomicText(
					ctx, database, tenant.schema, callIAMAtomicRefresh,
					tenant.schema, id, bytes.Repeat([]byte{0xd5 + index}, 16),
					bytes.Repeat([]byte{0xe5 + index}, 32),
				)
				results <- outcome
				errors <- err
			}()
		}
		outcomes := []string{<-results, <-results}
		for range 2 {
			if err := <-errors; err != nil {
				t.Fatalf("%s IAM refresh race: %v", tenant.schema, err)
			}
		}
		sort.Strings(outcomes)
		if outcomes[0] != "reused" || outcomes[1] != "rotated" {
			t.Fatalf("%s IAM refresh race outcomes = %v", tenant.schema, outcomes)
		}
		var familyState, sessionState string
		var activeTokens, successors int
		var tokenVersion int64
		mustQueryIAMTenantValue(t, ctx, database, tenant.schema, `
SELECT family.state, session.state, principal.token_version,
       (SELECT pg_catalog.count(*) FROM iam_refresh_tokens AS active
        WHERE active.family_id = family.id AND active.state = 'active'),
       (SELECT pg_catalog.count(*) FROM iam_refresh_tokens AS child
        WHERE child.parent_id = '93000000-0000-4000-8000-000000000001')
FROM iam_refresh_families AS family
JOIN iam_sessions AS session ON session.id = family.session_id
JOIN iam_users AS principal ON principal.id = family.user_id
WHERE family.id = '92000000-0000-4000-8000-000000000002'`,
			&familyState, &sessionState, &tokenVersion, &activeTokens, &successors)
		if familyState != "compromised" || sessionState != "revoked" ||
			tokenVersion != 2 || activeTokens != 0 || successors != 1 {
			t.Fatalf(
				"%s IAM reuse state = %s/%s/%d/%d/%d",
				tenant.schema, familyState, sessionState,
				tokenVersion, activeTokens, successors,
			)
		}
	}
}

func assertIAMRefreshCrashRollback(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatal(err)
	}
	var outcome string
	err = transaction.QueryRowContext(
		ctx, callIAMAtomicRefresh, schema,
		"93000000-0000-4000-8000-000000000009",
		bytes.Repeat([]byte{0xd8}, 16), bytes.Repeat([]byte{0xe4}, 32),
	).Scan(&outcome)
	if err != nil || outcome != "rotated" {
		_ = transaction.Rollback()
		_ = connection.Close()
		t.Fatalf("%s IAM refresh crash candidate = %q: %v", schema, outcome, err)
	}
	_ = transaction.Rollback()
	_ = connection.Close()
	var state string
	var revision int64
	var successorCount int
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT root.state, root.revision,
       (SELECT pg_catalog.count(*) FROM iam_refresh_tokens AS child
        WHERE child.parent_id = root.id)
FROM iam_refresh_tokens AS root
WHERE root.id = '93000000-0000-4000-8000-000000000001'`,
		&state, &revision, &successorCount)
	if state != "active" || revision != 1 || successorCount != 0 {
		t.Fatalf("%s IAM refresh crash rollback = %s/%d/%d",
			schema, state, revision, successorCount)
	}
}
