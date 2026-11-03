package baseline

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

const databaseOwnerRole = "harpy_owner"

func open(dsn, searchPath string) (*sql.DB, error) {
	return openSession(dsn, searchPath, false)
}

func openSession(dsn, searchPath string, assumeOwner bool) (*sql.DB, error) {
	config, err := sessionConfig(dsn, searchPath, assumeOwner)
	if err != nil {
		return nil, err
	}
	database := stdlib.OpenDB(*config)
	database.SetMaxOpenConns(1)
	return database, nil
}

func sessionConfig(dsn, searchPath string, assumeOwner bool) (*pgx.ConnConfig, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse baseline dsn: %w", err)
	}
	if config.RuntimeParams == nil {
		config.RuntimeParams = map[string]string{}
	}
	if searchPath != "" {
		config.RuntimeParams["search_path"] = searchPath
	}
	if assumeOwner {
		config.RuntimeParams["role"] = databaseOwnerRole
	}
	return config, nil
}

func ensureSchema(ctx context.Context, dsn, schema string) error {
	return ensureSchemaSession(ctx, dsn, schema, false)
}

func ensureOwnerSchema(ctx context.Context, dsn, schema string) error {
	return ensureSchemaSession(ctx, dsn, schema, true)
}

func ensureSchemaSession(ctx context.Context, dsn, schema string, assumeOwner bool) error {
	config, err := sessionConfig(dsn, "", assumeOwner)
	if err != nil {
		return err
	}
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("open baseline database: %w", err)
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+pgx.Identifier{schema}.Sanitize())
	if err != nil {
		return fmt.Errorf("create schema %s: %w", schema, err)
	}
	return nil
}

func validSchema(schema string) bool {
	if schema == "" || len(schema) > 63 {
		return false
	}
	for index, char := range schema {
		ok := char == '_' || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9' && index > 0)
		if !ok {
			return false
		}
	}
	return schema != tenantSchemaToken
}
