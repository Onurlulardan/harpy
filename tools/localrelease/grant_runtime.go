package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var runtimeReadinessTables = []string{
	"environment_identity",
	"bootstrap_receipts",
	"extension_registry",
	"migration_runs",
	"schema_contracts",
	"database_generation_manifests",
	"release_generations",
}

var readinessReaderRoles = []string{
	"harpy_runtime",
	"harpy_backup_reader",
}

func runGrantRuntime(ctx context.Context, arguments []string) error {
	path, err := parseSinglePath(arguments, "grant-runtime", "admin-dsn-file")
	if err != nil {
		return err
	}
	dsn, err := readSecret(path, "admin DSN")
	if err != nil {
		return err
	}
	admin, config, err := openAdmin(ctx, dsn)
	if err != nil {
		return err
	}
	defer admin.Close(context.Background())
	target, err := openTarget(ctx, config)
	if err != nil {
		return err
	}
	defer target.Close(context.Background())
	if _, err := readLiveDatabaseIdentity(ctx, target); err != nil {
		return err
	}
	if err := grantRuntimeReadiness(ctx, target); err != nil {
		return err
	}
	fmt.Println("readiness grants: verified")
	return nil
}

func grantRuntimeReadiness(ctx context.Context, target *pgx.Conn) error {
	if err := requireReadinessTables(ctx, target); err != nil {
		return err
	}
	tx, err := target.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SET LOCAL ROLE harpy_owner"); err != nil {
		return fmt.Errorf("assume readiness table owner: %w", err)
	}
	for _, role := range readinessReaderRoles {
		if _, err := tx.Exec(ctx, "GRANT CONNECT ON DATABASE harpy TO "+pgx.Identifier{role}.Sanitize()); err != nil {
			return fmt.Errorf("grant %s database connect: %w", role, err)
		}
		if _, err := tx.Exec(ctx, "GRANT USAGE ON SCHEMA platform TO "+pgx.Identifier{role}.Sanitize()); err != nil {
			return fmt.Errorf("grant %s schema usage: %w", role, err)
		}
		for _, table := range runtimeReadinessTables {
			statement := "GRANT SELECT ON TABLE " +
				pgx.Identifier{"platform", table}.Sanitize() + " TO " + pgx.Identifier{role}.Sanitize()
			if _, err := tx.Exec(ctx, statement); err != nil {
				return fmt.Errorf("grant %s readiness table %s: %w", role, table, err)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit runtime readiness grants: %w", err)
	}
	return verifyRuntimeReadinessGrants(ctx, target)
}

func grantExistingRuntimeReadiness(ctx context.Context, target *pgx.Conn) error {
	count := 0
	for _, table := range runtimeReadinessTables {
		var exists bool
		if err := target.QueryRow(ctx,
			"SELECT pg_catalog.to_regclass('platform.' || $1) IS NOT NULL", table,
		).Scan(&exists); err != nil {
			return fmt.Errorf("inspect existing readiness tables: %w", err)
		}
		if exists {
			count++
		}
	}
	if count == 0 {
		return nil
	}
	if count != len(runtimeReadinessTables) {
		return fmt.Errorf("runtime readiness table set is partially present")
	}
	return grantRuntimeReadiness(ctx, target)
}

func requireReadinessTables(ctx context.Context, target *pgx.Conn) error {
	for _, table := range runtimeReadinessTables {
		var exists, ownerExact bool
		err := target.QueryRow(ctx, `SELECT pg_catalog.to_regclass('platform.' || $1) IS NOT NULL,
EXISTS (SELECT 1 FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS schema ON schema.oid=relation.relnamespace
JOIN pg_catalog.pg_roles AS owner ON owner.oid=relation.relowner
WHERE schema.nspname='platform' AND relation.relname=$1
AND relation.relkind IN ('r','p') AND owner.rolname='harpy_owner')`, table).Scan(
			&exists, &ownerExact,
		)
		if err != nil {
			return fmt.Errorf("inspect readiness table %s: %w", table, err)
		}
		if !exists || !ownerExact {
			return fmt.Errorf("required readiness table platform.%s is missing or has wrong owner", table)
		}
	}
	return nil
}

func verifyRuntimeReadinessGrants(ctx context.Context, target *pgx.Conn) error {
	for _, role := range readinessReaderRoles {
		var connectAllowed, temporaryAllowed bool
		if err := target.QueryRow(ctx, `SELECT
pg_catalog.has_database_privilege($1,current_database(),'CONNECT'),
pg_catalog.has_database_privilege($1,current_database(),'TEMPORARY')`, role).Scan(
			&connectAllowed, &temporaryAllowed,
		); err != nil || !connectAllowed || temporaryAllowed {
			return fmt.Errorf("%s database ACL differs: %w", role, err)
		}
		var schemaUsage, schemaCreate bool
		if err := target.QueryRow(ctx, `SELECT
pg_catalog.has_schema_privilege($1,'platform','USAGE'),
pg_catalog.has_schema_privilege($1,'platform','CREATE')`, role).Scan(
			&schemaUsage, &schemaCreate,
		); err != nil || !schemaUsage || schemaCreate {
			return fmt.Errorf("%s platform schema ACL differs: %w", role, err)
		}
		for _, table := range runtimeReadinessTables {
			var selectAllowed, writeAllowed bool
			err := target.QueryRow(ctx, `SELECT
pg_catalog.has_table_privilege($1,'platform.' || $2,'SELECT'),
pg_catalog.has_table_privilege($1,'platform.' || $2,'INSERT,UPDATE,DELETE,TRUNCATE,REFERENCES,TRIGGER')`,
				role, table).Scan(&selectAllowed, &writeAllowed)
			if err != nil || !selectAllowed || writeAllowed {
				return fmt.Errorf("%s readiness ACL differs for platform.%s: %w", role, table, err)
			}
		}
	}
	return nil
}
