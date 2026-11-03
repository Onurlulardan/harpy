package baseline

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"testing/fstest"

	"github.com/jackc/pgx/v5"
	"github.com/pressly/goose/v3"
)

//go:embed sql/20261028000000_platform_baseline.fresh.sql
var platformScript string

//go:embed sql/20261028000001_tenant_baseline.fresh.sql
var tenantScript string

//go:embed sql/20261102000000_platform_catalog_removal_baseline.fresh.sql
var platformCatalogRemovalScript string

//go:embed sql/20261102000001_tenant_catalog_removal_baseline.fresh.sql
var tenantCatalogRemovalScript string

const (
	tenantSchemaToken             = "__tenant_schema__"
	platformVersion               = int64(20261028000000)
	platformCatalogRemovalVersion = int64(20261102000000)
	tenantVersion                 = int64(20261028000001)
	platformFile                  = "20261028000000_platform_baseline.fresh.sql"
	tenantFile                    = "20261028000001_tenant_baseline.fresh.sql"
	platformCatalogRemovalFile    = "20261102000000_platform_catalog_removal_baseline.fresh.sql"
	tenantCatalogRemovalFile      = "20261102000001_tenant_catalog_removal_baseline.fresh.sql"
)

type Result struct {
	Applied         bool
	Version         int64
	AppliedVersions []int64
}

func ApplyPlatform(ctx context.Context, dsn string) (Result, error) {
	if err := ensureSchema(ctx, dsn, "platform"); err != nil {
		return Result{}, err
	}
	result, err := apply(ctx, dsn, "platform, pg_catalog", platformFile, platformScript)
	if err != nil {
		return Result{}, err
	}
	cleanup, err := apply(
		ctx, dsn, "platform, pg_catalog",
		platformCatalogRemovalFile, platformCatalogRemovalScript,
	)
	if err != nil {
		return Result{}, fmt.Errorf("apply platform Catalog removal: %w", err)
	}
	executor, err := applyPlatformTenantExecutorBaseline(ctx, dsn)
	if err != nil {
		return Result{}, err
	}
	completion, err := applyPlatformTenantExecutorCompletion(ctx, dsn)
	if err != nil {
		return Result{}, err
	}
	return combineBaselineResults(result, cleanup, executor, completion), nil
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
	result, err := applyDB(ctx, database, tenantFile, script)
	if err != nil {
		return Result{}, err
	}
	cleanup := strings.ReplaceAll(tenantCatalogRemovalScript, tenantSchemaToken, schema)
	cleanupResult, err := applyDB(ctx, database, tenantCatalogRemovalFile, cleanup)
	if err != nil {
		return Result{}, fmt.Errorf("apply tenant Catalog removal: %w", err)
	}
	return combineBaselineResults(result, cleanupResult), nil
}

func ApplyTenant(ctx context.Context, dsn, schema string) (Result, error) {
	if !validSchema(schema) {
		return Result{}, fmt.Errorf("tenant schema name is invalid")
	}
	if err := ensureSchema(ctx, dsn, schema); err != nil {
		return Result{}, err
	}
	script := strings.ReplaceAll(tenantScript, tenantSchemaToken, schema)
	result, err := apply(ctx, dsn, schema+", pg_catalog", tenantFile, script)
	if err != nil {
		return Result{}, err
	}
	cleanup := strings.ReplaceAll(tenantCatalogRemovalScript, tenantSchemaToken, schema)
	cleanupResult, err := apply(
		ctx, dsn, schema+", pg_catalog", tenantCatalogRemovalFile, cleanup,
	)
	if err != nil {
		return Result{}, fmt.Errorf("apply tenant Catalog removal: %w", err)
	}
	return combineBaselineResults(result, cleanupResult), nil
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
	versionPrefix := strings.SplitN(name, "_", 2)[0]
	version, err := strconv.ParseInt(versionPrefix, 10, 64)
	if err != nil {
		return Result{}, fmt.Errorf("parse baseline migration version %s: %w", name, err)
	}
	return Result{Applied: len(applied) > 0, Version: version}, nil
}
