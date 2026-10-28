package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func openAdmin(ctx context.Context, dsn string) (*pgx.Conn, *pgx.ConnConfig, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("parse admin DSN: %w", err)
	}
	if config.Database != "postgres" {
		return nil, nil, fmt.Errorf("admin DSN must target maintenance database postgres")
	}
	if _, exists := config.RuntimeParams["role"]; exists {
		return nil, nil, fmt.Errorf("admin DSN must not set a role runtime parameter")
	}
	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, nil, fmt.Errorf("connect admin DSN: %w", err)
	}
	var database, sessionUser, currentUser string
	var superuser bool
	if err := connection.QueryRow(ctx, `SELECT current_database(), session_user,
current_user, rolsuper FROM pg_catalog.pg_roles WHERE rolname=session_user`).Scan(
		&database, &sessionUser, &currentUser, &superuser,
	); err != nil {
		connection.Close(ctx)
		return nil, nil, fmt.Errorf("verify admin connection: %w", err)
	}
	if database != "postgres" || sessionUser != currentUser || !superuser {
		connection.Close(ctx)
		return nil, nil, fmt.Errorf("admin DSN must be a direct PostgreSQL superuser session")
	}
	return connection, config, nil
}

func openTarget(
	ctx context.Context, adminConfig *pgx.ConnConfig,
) (*pgx.Conn, error) {
	config := adminConfig.Copy()
	config.Database = databaseName
	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("connect target database: %w", err)
	}
	return connection, nil
}
