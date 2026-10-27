//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMAuthSecurityRuntimeACL(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	tests := []struct {
		signature string
		allowed   bool
	}{
		{"iam_challenge_resolve_v1(name,uuid,bigint,smallint,text)", false},
		{"iam_recovery_consume_v1(name,uuid,bigint)", false},
		{"iam_reauth_binding_sha256_v1(uuid,uuid,uuid,bigint,bigint,text,uuid)", false},
		{"iam_reauth_context_sha256_v1(uuid,uuid,uuid,bytea,bigint,text,bytea)", false},
		{"iam_reauthentication_proof_apply_v1(name,uuid,text,uuid,bigint,bigint,bigint,bigint,bigint,boolean,boolean,bigint)", false},
		{"iam_reauthentication_start_v1(name,uuid,bytea,uuid,uuid,uuid,bigint,bigint,text,uuid,text,uuid,text,bytea,bytea,bytea,bigint,bytea,boolean)", true},
		{"iam_reauthentication_credential_v1(name,uuid,bytea,uuid,uuid,uuid,bigint,bigint,text,uuid,text,bytea,bigint,bytea,bytea)", true},
		{"iam_reauthentication_reject_v1(name,uuid,bigint,smallint,uuid,bigint,boolean,boolean)", true},
		{"iam_reauthentication_complete_v1(name,uuid,bytea,uuid,bigint,smallint,uuid,uuid,bigint,bigint,bigint,bigint,uuid,text,bigint,uuid,bigint,uuid,uuid,bigint,bigint,uuid,bigint,bigint,bigint,bigint,bytea,bytea,text,bigint,boolean,boolean,bigint,uuid,bytea,text,bytea)", true},
		{"iam_session_revoke_v1(name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint)", true},
		{"iam_recovery_credential_v1(name,bytea,text,bytea)", true},
		{"iam_recovery_consume_v2(name,bytea,text,bytea,uuid,bigint,uuid,bigint,uuid,bigint,bigint)", true},
	}
	for _, tenant := range iamFoundationTenants {
		for _, test := range tests {
			var allowed bool
			err := database.QueryRowContext(ctx, `
SELECT pg_catalog.has_function_privilege(
 'harpy_runtime',pg_catalog.to_regprocedure(
  pg_catalog.format('%I.%s',$1::text,$2::text)
 ),'EXECUTE')`, tenant.schema, test.signature).Scan(&allowed)
			if err != nil || allowed != test.allowed {
				t.Fatalf("%s %s privilege=%v want=%v: %v",
					tenant.schema, test.signature, allowed, test.allowed, err)
			}
		}
	}
}
