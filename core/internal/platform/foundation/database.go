package foundation

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func lookupDatabase(
	ctx context.Context, conn *pgx.Conn, name string,
) (bool, string, string, error) {
	var owner, sessionUser string
	err := conn.QueryRow(ctx, `SELECT pg_catalog.pg_get_userbyid(datdba), current_user
FROM pg_catalog.pg_database WHERE datname = $1`, name).Scan(&owner, &sessionUser)
	if err == pgx.ErrNoRows {
		if err := conn.QueryRow(ctx, "SELECT current_user").Scan(&sessionUser); err != nil {
			return false, "", "", err
		}
		return false, "", sessionUser, nil
	}
	if err != nil {
		return false, "", "", fmt.Errorf("lookup database %s: %w", name, err)
	}
	return true, owner, sessionUser, nil
}

func ensureDatabase(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	exists, owner, sessionUser, err := lookupDatabase(ctx, conn, cfg.Database)
	if err != nil {
		return failed("database", cfg, true, err)
	}
	if !exists {
		statement := fmt.Sprintf(
			"CREATE DATABASE %s OWNER %s TEMPLATE template0",
			ident(cfg.Database), ident(RoleOwner),
		)
		if _, err := conn.Exec(ctx, statement); err != nil {
			var pgErr *pgconn.PgError
			if asPgError(err, &pgErr) && pgErr.Code == "42P04" {
				return ensureExistingDatabase(ctx, conn, cfg)
			}
			return failed("database", cfg, true, fmt.Errorf("create database: %w", err))
		}
		return nil
	}
	return assignDatabaseOwner(cfg, owner, sessionUser, func() error {
		statement := fmt.Sprintf(
			"ALTER DATABASE %s OWNER TO %s", ident(cfg.Database), ident(RoleOwner),
		)
		if _, err := conn.Exec(ctx, statement); err != nil {
			return fmt.Errorf("assign database owner: %w", err)
		}
		return nil
	})
}

func ensureExistingDatabase(ctx context.Context, conn *pgx.Conn, cfg Config) error {
	exists, owner, sessionUser, err := lookupDatabase(ctx, conn, cfg.Database)
	if err != nil {
		return failed("database", cfg, true, err)
	}
	if !exists {
		return failed("database", cfg, true, fmt.Errorf("database %s disappeared", cfg.Database))
	}
	return assignDatabaseOwner(cfg, owner, sessionUser, func() error {
		statement := fmt.Sprintf(
			"ALTER DATABASE %s OWNER TO %s", ident(cfg.Database), ident(RoleOwner),
		)
		_, err := conn.Exec(ctx, statement)
		return err
	})
}

func assignDatabaseOwner(cfg Config, owner, sessionUser string, assign func() error) error {
	switch owner {
	case RoleOwner:
		return nil
	case sessionUser:
		if err := assign(); err != nil {
			return failed("database", cfg, true, err)
		}
		return nil
	default:
		return failed("database", cfg, false, fmt.Errorf(
			"database %s is owned by %s", cfg.Database, owner,
		))
	}
}

func asPgError(err error, target **pgconn.PgError) bool {
	for err != nil {
		if typed, ok := err.(*pgconn.PgError); ok {
			*target = typed
			return true
		}
		unwrap, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = unwrap.Unwrap()
	}
	return false
}
