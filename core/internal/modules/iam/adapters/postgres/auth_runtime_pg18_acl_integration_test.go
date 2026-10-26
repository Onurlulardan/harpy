//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMAuthRuntimeACL(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		var oldReader, oldRotate, reader, rotate bool
		err := database.QueryRowContext(ctx, `
SELECT
 pg_catalog.has_function_privilege(
  'harpy_runtime',pg_catalog.to_regprocedure(pg_catalog.format(
   '%I.iam_refresh_credential_v1(name,uuid,uuid,bytea,bigint,bytea)',$1::text
  )),'EXECUTE'),
 pg_catalog.has_function_privilege(
  'harpy_runtime',pg_catalog.to_regprocedure(pg_catalog.format(
   '%I.iam_refresh_rotate_v2(name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,uuid,bigint,uuid,bigint,uuid,bytea,text,bytea)',$1::text
  )),'EXECUTE'),
 pg_catalog.has_function_privilege(
  'harpy_runtime',pg_catalog.to_regprocedure(pg_catalog.format(
   '%I.iam_refresh_credential_v2(name,uuid,uuid,bytea,bigint,bytea)',$1::text
  )),'EXECUTE'),
 pg_catalog.has_function_privilege(
  'harpy_runtime',pg_catalog.to_regprocedure(pg_catalog.format(
   '%I.iam_refresh_rotate_v3(name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,uuid,bigint,uuid,bigint,uuid,bytea,text,bytea)',$1::text
  )),'EXECUTE')`, tenant.schema).Scan(
			&oldReader, &oldRotate, &reader, &rotate,
		)
		if err != nil || oldReader || oldRotate || !reader || !rotate {
			t.Fatalf("%s IAM auth ACL old=%v/%v current=%v/%v: %v",
				tenant.schema, oldReader, oldRotate, reader, rotate, err)
		}
	}
}
