package main

import (
	"context"
	"fmt"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
)

func resolveFunctionTransferIdentity(
	ctx context.Context,
	transaction pgx.Tx,
	object cutoverv1.AuthorityAdoptionObject,
) (string, error) {
	var identity string
	err := transaction.QueryRow(ctx, `SELECT procedure_row.oid::regprocedure::text
FROM pg_catalog.pg_proc AS procedure_row
JOIN pg_catalog.pg_namespace AS namespace
  ON namespace.oid=procedure_row.pronamespace
WHERE namespace.nspname=$1 AND procedure_row.proname=$2
  AND pg_catalog.replace(pg_catalog.oidvectortypes(
      procedure_row.proargtypes),' ','')=$3`,
		object.SchemaName, object.ObjectName, object.IdentityArguments,
	).Scan(&identity)
	if err != nil || identity == "" {
		return "", fmt.Errorf("authority function transfer identity is missing")
	}
	return identity, nil
}
