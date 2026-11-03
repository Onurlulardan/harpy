package foundation

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5"
)

func connect(ctx context.Context, dsn string) (*pgx.Conn, error) {
	parsed, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	if parsed.RuntimeParams == nil {
		parsed.RuntimeParams = map[string]string{}
	}
	parsed.RuntimeParams["application_name"] = "harpy-initialize"
	return pgx.ConnectConfig(ctx, parsed)
}

func databaseDSN(adminDSN, database string) (string, error) {
	parsed, err := url.Parse(adminDSN)
	if err != nil {
		return "", err
	}
	parsed.Path = "/" + database
	return parsed.String(), nil
}

func userDSN(adminDSN, user, password, database string) (string, error) {
	parsed, err := url.Parse(adminDSN)
	if err != nil {
		return "", err
	}
	parsed.User = url.UserPassword(user, password)
	parsed.Path = "/" + database
	return parsed.String(), nil
}

func ident(value string) string {
	return pgx.Identifier{value}.Sanitize()
}

func currentUser(ctx context.Context, conn *pgx.Conn) (string, error) {
	var name string
	if err := conn.QueryRow(ctx, "SELECT current_user").Scan(&name); err != nil {
		return "", fmt.Errorf("read current user: %w", err)
	}
	return name, nil
}
