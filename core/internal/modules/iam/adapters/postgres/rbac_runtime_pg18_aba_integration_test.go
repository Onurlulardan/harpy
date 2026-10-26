//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"testing"
)

const iamRBACABABindingID = "98500000-0000-4000-8000-000000000001"

const writeIAMRBACUserBinding = `
SELECT pg_catalog.concat_ws('|', result_id::text, result_revision::text,
    authorization_clock::text, replayed::text)
FROM iam_rbac_user_binding_%s_v1(
    pg_catalog.current_schema()::name, $1::uuid, $2::bytea, $3::uuid,
    '51000000-0000-4000-8000-000000000002',
    '93000000-0000-4000-8000-000000000001', $4::bigint,
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001', 1, 2
)`

func assertIAMRBACBindingABA(t *testing.T, ctx context.Context,
	database *sql.DB, schema string) {
	t.Helper()
	tests := []struct {
		op        string
		operation string
		revision  int64
		want      string
	}{
		{"grant", "98600000-0000-4000-8000-000000000001", 0, iamRBACABABindingID + "|1|2|false"},
		{"revoke", "98600000-0000-4000-8000-000000000002", 1, iamRBACABABindingID + "|2|3|false"},
		{"grant", "98600000-0000-4000-8000-000000000003", 2, iamRBACABABindingID + "|3|4|false"},
	}
	for index, test := range tests {
		hash := bytes.Repeat([]byte{byte(0x50 + index)}, 32)
		result := mustRunIAMAtomicText(t, ctx, database, schema,
			fmt.Sprintf(writeIAMRBACUserBinding, test.op),
			test.operation, hash, iamRBACABABindingID, test.revision)
		if result != test.want {
			t.Fatalf("%s ABA result = %s", test.op, result)
		}
	}
	staleHash := bytes.Repeat([]byte{0x59}, 32)
	_, err := runIAMAtomicText(ctx, database, schema,
		fmt.Sprintf(writeIAMRBACUserBinding, "revoke"),
		"98600000-0000-4000-8000-000000000004", staleHash, iamRBACABABindingID, int64(1))
	if err == nil {
		t.Fatal("stale ABA revision was accepted")
	}
	var state string
	var revision, targetVersion int64
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT binding.state, binding.revision, principal.authorization_version
FROM iam_user_permission_bindings AS binding
JOIN iam_users AS principal ON principal.id=binding.user_id
WHERE binding.id='98500000-0000-4000-8000-000000000001'`,
		&state, &revision, &targetVersion)
	if state != "active" || revision != 3 || targetVersion != 4 ||
		iamRBACClock(t, ctx, database, schema) != 4 {
		t.Fatalf("ABA state = %s/%d/%d", state, revision, targetVersion)
	}
}
