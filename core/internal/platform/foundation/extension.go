package foundation

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func ensureExtension(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	var available bool
	err := conn.QueryRow(ctx, `SELECT EXISTS(
SELECT 1 FROM pg_catalog.pg_available_extensions WHERE name = $1)`,
		VectorExtension).Scan(&available)
	if err != nil {
		return failed("extension", cfg, true, err)
	}
	if !available {
		return failed("extension", cfg, false, fmt.Errorf(
			"extension %s is not available", VectorExtension,
		))
	}
	var schemaName, owner string
	err = conn.QueryRow(ctx, `SELECT namespace.nspname, pg_catalog.pg_get_userbyid(extension.extowner)
FROM pg_catalog.pg_extension AS extension
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = extension.extnamespace
WHERE extension.extname = $1`, VectorExtension).Scan(&schemaName, &owner)
	if err == pgx.ErrNoRows {
		statement := fmt.Sprintf(
			"CREATE EXTENSION %s WITH SCHEMA %s", ident(VectorExtension), ident(PlatformSchema),
		)
		if _, err := conn.Exec(ctx, statement); err != nil {
			return failed("extension", cfg, true, fmt.Errorf("create extension: %w", err))
		}
		return nil
	}
	if err != nil {
		return failed("extension", cfg, true, err)
	}
	if schemaName != PlatformSchema {
		return failed("extension", cfg, false, fmt.Errorf(
			"extension %s is in schema %s", VectorExtension, schemaName,
		))
	}
	sessionUser, err := currentUser(ctx, conn)
	if err != nil {
		return failed("extension", cfg, true, err)
	}
	if owner != RoleOwner && owner != sessionUser {
		return failed("extension", cfg, false, fmt.Errorf(
			"extension %s is owned by %s", VectorExtension, owner,
		))
	}
	return nil
}
