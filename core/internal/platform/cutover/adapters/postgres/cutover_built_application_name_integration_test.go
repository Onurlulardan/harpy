//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func assertBuiltRoleApplicationName(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	role, applicationName string,
) {
	t.Helper()
	configuration, err := pgx.ParseConfig(database.dsn)
	if err != nil {
		t.Fatal(err)
	}
	configuration.User = role
	configuration.Password = builtCutoverPassword
	configuration.RuntimeParams["application_name"] = applicationName
	connection, err := pgx.ConnectConfig(ctx, configuration)
	if err != nil {
		t.Fatalf("connect %s with application_name: %v", role, err)
	}
	defer connection.Close(ctx)
	var actualRole, actualApplication string
	err = connection.QueryRow(ctx,
		"SELECT SESSION_USER, current_setting('application_name')").Scan(
		&actualRole, &actualApplication,
	)
	if err != nil || actualRole != role || actualApplication != applicationName {
		t.Fatalf("application session = %s %q, want %s %q, err=%v",
			actualRole, actualApplication, role, applicationName, err)
	}
}
