package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type DeployDatabase struct {
	inner            *sql.DB
	tenantConnection *pgx.ConnConfig
	name             string
	closeOnce        sync.Once
	closeErr         error
}

func NewDeployDatabase(settings config.Postgres) (*DeployDatabase, error) {
	connectionConfig, err := prepareDeployConnection(settings)
	if err != nil {
		return nil, err
	}
	database := stdlib.OpenDB(*connectionConfig)
	database.SetMaxOpenConns(int(settings.MaxConnections()))
	database.SetMaxIdleConns(int(settings.MinConnections()))
	database.SetConnMaxLifetime(settings.MaxConnectionLifetime())
	database.SetConnMaxIdleTime(settings.MaxConnectionIdleTime())
	return &DeployDatabase{
		inner: database, tenantConnection: connectionConfig.Copy(),
		name: connectionConfig.Database,
	}, nil
}

func (database *DeployDatabase) OpenTenantDatabase() (*sql.DB, error) {
	if database == nil || database.tenantConnection == nil {
		return nil, fmt.Errorf("tenant migration database configuration is absent")
	}
	connectionConfig := database.tenantConnection.Copy()
	connectionConfig.RuntimeParams["application_name"] =
		"harpy-v2:core-migrator-tenant"
	return openOneShotTenantDatabase(connectionConfig), nil
}

func openOneShotTenantDatabase(connectionConfig *pgx.ConnConfig) *sql.DB {
	var connectionOpened atomic.Bool
	var backendPID atomic.Uint32
	inner := stdlib.OpenDB(
		*connectionConfig,
		stdlib.OptionBeforeConnect(func(_ context.Context, _ *pgx.ConnConfig) error {
			if !connectionOpened.CompareAndSwap(false, true) {
				return fmt.Errorf("tenant migration session reconnect is forbidden")
			}
			return nil
		}),
		stdlib.OptionAfterConnect(func(_ context.Context, connection *pgx.Conn) error {
			pid := connection.PgConn().PID()
			if pid == 0 || !backendPID.CompareAndSwap(0, pid) {
				return fmt.Errorf("tenant migration backend identity is invalid")
			}
			return nil
		}),
	)
	inner.SetMaxOpenConns(1)
	inner.SetMaxIdleConns(1)
	return inner
}

func prepareDeployConnection(settings config.Postgres) (*pgx.ConnConfig, error) {
	connectionConfig, err := pgx.ParseConfig(settings.DSN().Value())
	if err != nil || connectionConfig.Database == "" {
		return nil, fmt.Errorf("migrator database configuration is invalid")
	}
	connectionConfig.ConnectTimeout = settings.ConnectTimeout()
	connectionConfig.RuntimeParams["role"] = "harpy_owner"
	connectionConfig.RuntimeParams["search_path"] = "pg_catalog"
	connectionConfig.RuntimeParams["application_name"] = "harpy-v2:core-migrator-deploy"
	return connectionConfig, nil
}

func (database *DeployDatabase) Name() string {
	if database == nil {
		return ""
	}
	return database.name
}

func (database *DeployDatabase) SQL() *sql.DB {
	if database == nil {
		return nil
	}
	return database.inner
}

func (database *DeployDatabase) Close() error {
	if database == nil {
		return nil
	}
	database.closeOnce.Do(func() { database.closeErr = database.inner.Close() })
	if database.closeErr != nil {
		return fmt.Errorf("close migrator database")
	}
	return nil
}
