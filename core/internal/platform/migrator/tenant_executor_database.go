package migrator

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type TenantExecutorDatabase struct {
	inner     *sql.DB
	tenant    *pgx.ConnConfig
	name      string
	closeOnce sync.Once
	closeErr  error
}

func NewTenantExecutorDatabase(
	settings config.Postgres,
) (*TenantExecutorDatabase, error) {
	connection, err := pgx.ParseConfig(settings.DSN().Value())
	if err != nil || connection.Database == "" {
		return nil, fmt.Errorf("tenant migration database configuration is invalid")
	}
	connection.ConnectTimeout = settings.ConnectTimeout()
	connection.RuntimeParams["role"] = "harpy_owner"
	connection.RuntimeParams["search_path"] = "pg_catalog"
	connection.RuntimeParams["application_name"] =
		"harpy-v2:core-tenant-migration-executor"
	inner := stdlib.OpenDB(*connection)
	inner.SetMaxOpenConns(int(settings.MaxConnections()))
	inner.SetMaxIdleConns(int(settings.MinConnections()))
	inner.SetConnMaxLifetime(settings.MaxConnectionLifetime())
	inner.SetConnMaxIdleTime(settings.MaxConnectionIdleTime())
	return &TenantExecutorDatabase{inner: inner, tenant: connection.Copy(),
		name: connection.Database}, nil
}

func (database *TenantExecutorDatabase) OpenTenant() (*sql.DB, error) {
	if database == nil || database.tenant == nil {
		return nil, fmt.Errorf("tenant migration database configuration is absent")
	}
	connection := database.tenant.Copy()
	connection.RuntimeParams["application_name"] =
		"harpy-v2:core-tenant-migration-target"
	return openOneShotTenantDatabase(connection), nil
}

func (database *TenantExecutorDatabase) SQL() *sql.DB { return database.inner }
func (database *TenantExecutorDatabase) Name() string { return database.name }

func (database *TenantExecutorDatabase) Close() error {
	if database == nil {
		return nil
	}
	database.closeOnce.Do(func() { database.closeErr = database.inner.Close() })
	return database.closeErr
}
