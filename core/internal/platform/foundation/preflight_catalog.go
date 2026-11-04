package foundation

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func preflightSchemas(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	rows, err := conn.Query(ctx, `SELECT nspname, pg_catalog.pg_get_userbyid(nspowner)
FROM pg_catalog.pg_namespace
WHERE nspname NOT IN ('pg_catalog', 'information_schema')
  AND nspname NOT LIKE 'pg\_%' ESCAPE '\'
ORDER BY nspname`)
	if err != nil {
		return failed("preflight", cfg, true, err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, owner string
		if err := rows.Scan(&name, &owner); err != nil {
			return failed("preflight", cfg, true, err)
		}
		if !AllowedSchema(name) {
			return failed("preflight", cfg, false, fmt.Errorf("unexpected schema %s", name))
		}
		if name == PlatformSchema && owner != RoleOwner {
			return failed("preflight", cfg, false, fmt.Errorf(
				"schema %s is owned by %s", name, owner,
			))
		}
	}
	if err := rows.Err(); err != nil {
		return failed("preflight", cfg, true, err)
	}
	return nil
}

func preflightExtensions(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	rows, err := conn.Query(ctx, `SELECT extension.extname, namespace.nspname
FROM pg_catalog.pg_extension AS extension
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = extension.extnamespace
ORDER BY extension.extname`)
	if err != nil {
		return failed("preflight", cfg, true, err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, schema string
		if err := rows.Scan(&name, &schema); err != nil {
			return failed("preflight", cfg, true, err)
		}
		if !AllowedExtension(name) {
			return failed("preflight", cfg, false, fmt.Errorf("unexpected extension %s", name))
		}
		if name == VectorExtension && schema != PlatformSchema {
			return failed("preflight", cfg, false, fmt.Errorf(
				"extension %s is in schema %s", name, schema,
			))
		}
	}
	return failed("preflight", cfg, true, rows.Err())
}

func preflightRelations(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	rows, err := conn.Query(ctx, `SELECT namespace.nspname, relation.relname
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
WHERE namespace.nspname NOT IN ('pg_catalog', 'information_schema')
  AND namespace.nspname NOT LIKE 'pg\_%' ESCAPE '\'
  AND relation.relkind IN ('r', 'p', 'v', 'm', 'S', 'f')
  AND relation.relpersistence <> 't'
ORDER BY 1, 2`)
	if err != nil {
		return failed("preflight", cfg, true, err)
	}
	defer rows.Close()
	var outside []string
	managed := 0
	gooseHistory := false
	for rows.Next() {
		var schemaName, relationName string
		if err = rows.Scan(&schemaName, &relationName); err != nil {
			return failed("preflight", cfg, true, err)
		}
		if schemaName == PlatformSchema || tenantSchemaPattern.MatchString(schemaName) {
			managed++
			if schemaName == PlatformSchema && relationName == "goose_db_version" {
				gooseHistory = true
			}
			continue
		}
		if len(outside) < 5 {
			outside = append(outside, schemaName+"."+relationName)
		}
	}
	if err = rows.Err(); err != nil {
		return failed("preflight", cfg, true, err)
	}
	if len(outside) > 0 {
		return failed("preflight", cfg, false, fmt.Errorf(
			"database is not an empty foundation target: %s", strings.Join(outside, ", "),
		))
	}
	if managed > 0 && !gooseHistory {
		return failed("preflight", cfg, false, fmt.Errorf(
			"database has relations without goose history",
		))
	}
	return nil
}

func preflightSchemaPrivileges(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	var exists bool
	err := conn.QueryRow(ctx, `SELECT EXISTS(
SELECT 1 FROM pg_catalog.pg_namespace WHERE nspname = $1)`, PlatformSchema).Scan(&exists)
	if err != nil {
		return failed("preflight", cfg, true, err)
	}
	if !exists {
		return nil
	}
	for _, role := range []string{RoleAPI, RoleWorker, RoleGoose} {
		var roleExists bool
		err = conn.QueryRow(ctx, `SELECT EXISTS(
SELECT 1 FROM pg_catalog.pg_roles WHERE rolname = $1)`, role).Scan(&roleExists)
		if err != nil {
			return failed("preflight", cfg, true, err)
		}
		if !roleExists {
			continue
		}
		var create bool
		err = conn.QueryRow(ctx, `SELECT pg_catalog.has_schema_privilege($1, $2, 'CREATE')`,
			role, PlatformSchema).Scan(&create)
		if err != nil {
			return failed("preflight", cfg, true, err)
		}
		if create {
			return failed("preflight", cfg, false, fmt.Errorf(
				"role %s can create objects in %s", role, PlatformSchema,
			))
		}
	}
	return nil
}
