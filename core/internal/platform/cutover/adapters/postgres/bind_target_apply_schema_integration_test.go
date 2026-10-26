//go:build integration

package postgres

import (
	"context"
	"testing"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/jackc/pgx/v5"
)

func verifyTargetApplySchemaBinding(
	t *testing.T,
	ctx context.Context,
	database cutoverIntegrationDatabase,
	adopter *pgx.Conn,
) {
	t.Helper()
	procedure := readApplyProcedureIdentity(t, ctx, database.admin)
	valid := readApplyTargetIdentity(t, ctx, database.admin,
		"platform", "integration_business_rows")
	transaction, err := adopter.Begin(ctx)
	if err != nil {
		t.Fatalf("begin valid target apply binding: %v", err)
	}
	if err = bindTargetApplySchema(ctx, transaction,
		cutoverv1.RelationMapping{Target: valid}, procedure); err != nil {
		_ = transaction.Rollback(ctx)
		t.Fatalf("bind exact target apply schema: %v", err)
	}
	var searchPath string
	if err = transaction.QueryRow(ctx,
		"SELECT pg_catalog.current_setting('search_path')",
	).Scan(&searchPath); err != nil || searchPath != "platform,pg_catalog" {
		_ = transaction.Rollback(ctx)
		t.Fatalf("bound target search_path=%q err=%v", searchPath, err)
	}
	_ = transaction.Rollback(ctx)

	mismatch := readApplyTargetIdentity(t, ctx, database.admin,
		"legacy_source", "allowed_rows")
	transaction, err = adopter.Begin(ctx)
	if err != nil {
		t.Fatalf("begin mismatched target apply binding: %v", err)
	}
	if err = bindTargetApplySchema(ctx, transaction,
		cutoverv1.RelationMapping{Target: mismatch}, procedure); err == nil {
		_ = transaction.Rollback(ctx)
		t.Fatal("cross-schema apply procedure binding succeeded")
	}
	_ = transaction.Rollback(ctx)
}

func readApplyProcedureIdentity(
	t *testing.T, ctx context.Context, admin *pgx.Conn,
) preparedBatch {
	t.Helper()
	value := preparedBatch{procedureSignature: "platform.cutover_probe_apply_v1(" +
		"uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)"}
	if err := admin.QueryRow(ctx,
		"SELECT $1::text::regprocedure::oid::bigint", value.procedureSignature,
	).Scan(&value.procedureOID); err != nil {
		t.Fatalf("read apply procedure identity: %v", err)
	}
	return value
}

func readApplyTargetIdentity(
	t *testing.T,
	ctx context.Context,
	admin *pgx.Conn,
	schema string,
	relation string,
) cutoverv1.RelationIdentity {
	t.Helper()
	value := cutoverv1.RelationIdentity{SchemaName: schema, RelationName: relation}
	if err := admin.QueryRow(ctx, `SELECT namespace.oid::bigint, relation.oid::bigint
FROM pg_catalog.pg_namespace AS namespace
JOIN pg_catalog.pg_class AS relation ON relation.relnamespace=namespace.oid
WHERE namespace.nspname=$1 AND relation.relname=$2`, schema, relation).Scan(
		&value.SchemaOID, &value.RelationOID,
	); err != nil {
		t.Fatalf("read apply target identity: %v", err)
	}
	return value
}
