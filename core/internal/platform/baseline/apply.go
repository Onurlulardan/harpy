package baseline

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
	"testing/fstest"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed sql/20261028000000_platform_baseline.sql
var platformScript string

//go:embed sql/20261028000000_tenant_baseline.sql
var tenantScript string

const (
	tenantSchemaToken = "__tenant_schema__"
	baselineVersion   = int64(20261028000000)
	platformFile      = "20261028000000_platform_baseline.sql"
	tenantFile        = "20261028000000_tenant_baseline.sql"
)

type Result struct {
	Applied bool
	Version int64
}

func ApplyPlatform(ctx context.Context, dsn string) (Result, error) {
	if err := ensureSchema(ctx, dsn, "platform"); err != nil {
		return Result{}, err
	}
	return apply(ctx, dsn, "platform, pg_catalog", platformFile, platformScript)
}

func ApplyTenantDB(ctx context.Context, database *sql.DB, schema string) (Result, error) {
	if !validSchema(schema) {
		return Result{}, fmt.Errorf("tenant schema name is invalid")
	}
	database.SetMaxOpenConns(1)
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err := database.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+quoted); err != nil {
		return Result{}, fmt.Errorf("create schema %s: %w", schema, err)
	}
	if _, err := database.ExecContext(ctx, "SET search_path TO "+quoted+", pg_catalog"); err != nil {
		return Result{}, fmt.Errorf("set tenant search path: %w", err)
	}
	script := strings.ReplaceAll(tenantScript, tenantSchemaToken, schema)
	return applyDB(ctx, database, tenantFile, script)
}

func ApplyTenant(ctx context.Context, dsn, schema string) (Result, error) {
	if !validSchema(schema) {
		return Result{}, fmt.Errorf("tenant schema name is invalid")
	}
	if err := ensureSchema(ctx, dsn, schema); err != nil {
		return Result{}, err
	}
	script := strings.ReplaceAll(tenantScript, tenantSchemaToken, schema)
	return apply(ctx, dsn, schema+", pg_catalog", tenantFile, script)
}

func apply(ctx context.Context, dsn, searchPath, name, script string) (Result, error) {
	database, err := open(dsn, searchPath)
	if err != nil {
		return Result{}, err
	}
	defer database.Close()
	return applyDB(ctx, database, name, script)
}

func applyDB(ctx context.Context, database *sql.DB, name, script string) (Result, error) {
	provider, err := goose.NewProvider(
		goose.DialectPostgres, database, fstest.MapFS{
			name: &fstest.MapFile{Data: []byte(script)},
		},
		goose.WithAllowOutofOrder(false),
		goose.WithDisableGlobalRegistry(true),
	)
	if err != nil {
		return Result{}, fmt.Errorf("create baseline provider: %w", err)
	}
	applied, err := provider.Up(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("apply baseline: %w", err)
	}
	return Result{Applied: len(applied) > 0, Version: baselineVersion}, nil
}

func open(dsn, searchPath string) (*sql.DB, error) {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse baseline dsn: %w", err)
	}
	config.RuntimeParams["search_path"] = searchPath
	database := stdlib.OpenDB(*config)
	database.SetMaxOpenConns(1)
	return database, nil
}

func ensureSchema(ctx context.Context, dsn, schema string) error {
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("parse baseline dsn: %w", err)
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
