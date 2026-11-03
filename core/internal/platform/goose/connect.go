package goose

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func openGoose(ctx context.Context, dsn string) (*sql.DB, error) {
	parsed, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse goose dsn: %w", err)
	}
	if parsed.RuntimeParams == nil {
		parsed.RuntimeParams = map[string]string{}
	}
	parsed.RuntimeParams["application_name"] = "harpy-goose"
	database := stdlib.OpenDB(*parsed)
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	if err = database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("open goose database: %w", err)
	}
	return database, nil
}

func userDSN(adminDSN, user, password, database string) (string, error) {
	parsed, err := url.Parse(adminDSN)
	if err != nil {
		return "", fmt.Errorf("parse admin dsn: %w", err)
	}
	parsed.User = url.UserPassword(user, password)
	parsed.Path = "/" + database
	return parsed.String(), nil
}

func gooseDSN(cfg foundation.Config) (string, error) {
	return userDSN(cfg.AdminDSN, foundation.RoleGoose, cfg.GoosePassword, cfg.Database)
}
