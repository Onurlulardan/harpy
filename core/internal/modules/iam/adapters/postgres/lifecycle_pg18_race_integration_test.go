//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"sort"
	"testing"
)

func assertIAMLifecycleSameOperationRace(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	actor := readIAMIdentityActor(t, ctx, database, tenant.schema, iamUserOneID)
	arguments := lifecycleRevokeArguments(tenant, iamLifecycleRaceOperation,
		bytes.Repeat([]byte{0x75}, 32), iamLifecycleRaceSession, 1, actor)
	type outcome struct {
		value string
		err   error
	}
	results := make(chan outcome, 2)
	for attempt := 0; attempt < 2; attempt++ {
		go func() {
			value, err := runIAMAtomicText(ctx, database, tenant.schema,
				revokeIAMLifecycleSession, arguments...)
			results <- outcome{value: value, err: err}
		}()
	}
	values := make([]string, 0, 2)
	for attempt := 0; attempt < 2; attempt++ {
		result := <-results
		if result.err != nil {
			t.Fatalf("%s lifecycle race: %v", tenant.schema, result.err)
		}
		values = append(values, result.value)
	}
	sort.Strings(values)
	want := []string{
		iamLifecycleRaceSession + "|2|false",
		iamLifecycleRaceSession + "|2|true",
	}
	if values[0] != want[0] || values[1] != want[1] {
		t.Fatalf("%s lifecycle race=%v want=%v", tenant.schema, values, want)
	}
	assertIAMLifecycleState(t, ctx, database, tenant.schema,
		iamLifecycleRaceSession, "revoked", 2)
	assertIAMLifecycleReceiptCount(t, ctx, database, tenant.schema,
		iamLifecycleRaceOperation, 1)
}
