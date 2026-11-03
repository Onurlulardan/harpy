package foundation

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func ensurePlatformSchema(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	var owner string
	err := conn.QueryRow(ctx, `SELECT pg_catalog.pg_get_userbyid(nspowner)
FROM pg_catalog.pg_namespace WHERE nspname = $1`, PlatformSchema).Scan(&owner)
	if err == pgx.ErrNoRows {
		statement := fmt.Sprintf(
			"CREATE SCHEMA %s AUTHORIZATION %s", ident(PlatformSchema), ident(RoleOwner),
		)
		if _, err := conn.Exec(ctx, statement); err != nil {
			return failed("schema", cfg, true, fmt.Errorf("create platform schema: %w", err))
		}
		return nil
	}
	if err != nil {
		return failed("schema", cfg, true, err)
	}
	if owner != RoleOwner {
		return failed("schema", cfg, false, fmt.Errorf(
			"schema %s is owned by %s", PlatformSchema, owner,
		))
	}
	return nil
}

func assertNoApplicationRelations(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	var relations int
	err := conn.QueryRow(ctx, `SELECT count(*)
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
WHERE namespace.nspname NOT IN ('pg_catalog', 'information_schema')
  AND namespace.nspname NOT LIKE 'pg\_%' ESCAPE '\'
  AND relation.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')
  AND relation.relpersistence <> 't'`).Scan(&relations)
	if err != nil {
		return failed("catalog", cfg, true, err)
	}
	if relations != 0 {
		return failed("catalog", cfg, false, fmt.Errorf(
			"foundation created %d application relations", relations,
		))
	}
	var goose int
	err = conn.QueryRow(ctx, `SELECT count(*)
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
WHERE relation.relname = 'goose_db_version'`).Scan(&goose)
	if err != nil {
		return failed("catalog", cfg, true, err)
	}
	if goose != 0 {
		return failed("catalog", cfg, false, fmt.Errorf("goose history exists after foundation"))
	}
	return nil
}
