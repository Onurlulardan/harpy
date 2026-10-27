package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/jackc/pgx/v5"
)

func openConnection(
	ctx context.Context,
	settings config.PrivilegedPostgres,
	applicationName string,
) (*pgx.Conn, error) {
	connectionConfig, err := pgx.ParseConfig(settings.DSN().Value())
	if err != nil || connectionConfig.Database == "" {
		return nil, fmt.Errorf("cutover PostgreSQL configuration is invalid")
	}
	connectionConfig.ConnectTimeout = settings.ConnectTimeout()
	connectionConfig.RuntimeParams["application_name"] = applicationName
	connectionConfig.RuntimeParams["search_path"] = "pg_catalog"
	connection, err := pgx.ConnectConfig(ctx, connectionConfig)
	if err != nil {
		return nil, fmt.Errorf("open cutover PostgreSQL session")
	}
	return connection, nil
}

func closeConnections(ctx context.Context, connections []*pgx.Conn) error {
	var closeFailed bool
	for _, connection := range connections {
		if connection != nil && connection.Close(ctx) != nil {
			closeFailed = true
		}
	}
	if closeFailed {
		return fmt.Errorf("close cutover PostgreSQL sessions")
	}
	return nil
}
