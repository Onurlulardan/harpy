//go:build integration

package postgres

import (
	"context"
	"io"
	"testing"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func prepareSchemaExecutorRoles(
	t *testing.T,
	ctx context.Context,
	container *tcpostgres.PostgresContainer,
) {
	t.Helper()
	statements := []string{
		"CREATE ROLE harpy_owner NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB " +
			"NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_wrong_owner NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB " +
			"NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_cutover_authority_owner NOLOGIN NOINHERIT NOSUPERUSER " +
			"NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT -1 " +
			"PASSWORD NULL VALID UNTIL 'infinity'",
		"CREATE ROLE harpy_runtime NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB " +
			"NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_forms_record_reader NOLOGIN NOINHERIT NOSUPERUSER " +
			"NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_forms_record_writer NOLOGIN NOINHERIT NOSUPERUSER " +
			"NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_v2_runtime LOGIN NOINHERIT NOSUPERUSER NOCREATEDB " +
			"NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_migrator LOGIN NOINHERIT NOSUPERUSER NOCREATEDB " +
			"NOCREATEROLE NOREPLICATION NOBYPASSRLS",
		"CREATE ROLE harpy_cluster_admin NOLOGIN NOINHERIT NOSUPERUSER CREATEDB " +
			"CREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 1",
		"CREATE ROLE harpy_extension_admin NOLOGIN NOINHERIT SUPERUSER NOCREATEDB " +
			"NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 1",
		"CREATE ROLE harpy_adopter NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB " +
			"NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0",
		"CREATE ROLE harpy_cutover_decoder NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB " +
			"NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0",
		"CREATE ROLE harpy_cutover_finalizer NOLOGIN NOINHERIT NOSUPERUSER NOCREATEDB " +
			"NOCREATEROLE NOREPLICATION NOBYPASSRLS CONNECTION LIMIT 0",
		"GRANT harpy_wrong_owner TO harpy_owner " +
			"WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"GRANT harpy_runtime TO harpy_v2_runtime " +
			"WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"GRANT harpy_forms_record_reader TO harpy_v2_runtime " +
			"WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"GRANT harpy_forms_record_writer TO harpy_v2_runtime " +
			"WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"GRANT harpy_owner TO harpy_migrator " +
			"WITH ADMIN FALSE, INHERIT FALSE, SET TRUE",
		"GRANT pg_signal_backend TO harpy_owner",
	}
	for _, statement := range statements {
		runSchemaExecutorBootstrapSQL(t, ctx, container, statement)
	}
	runSchemaExecutorBootstrapSQL(t, ctx, container,
		"ALTER ROLE harpy_migrator PASSWORD '"+schemaExecutorMigratorPassword+"'")
	runSchemaExecutorBootstrapSQL(t, ctx, container,
		"CREATE DATABASE harpy_adapter OWNER harpy_owner")
	runSchemaExecutorBootstrapSQL(t, ctx, container,
		"REVOKE CONNECT, TEMPORARY ON DATABASE harpy_adapter FROM PUBLIC")
	runSchemaExecutorBootstrapSQL(t, ctx, container,
		"GRANT CONNECT ON DATABASE harpy_adapter TO harpy_owner, harpy_runtime, "+
			"harpy_v2_runtime, harpy_migrator, harpy_cluster_admin, "+
			"harpy_extension_admin")
}

func runSchemaExecutorBootstrapSQL(
	t *testing.T,
	ctx context.Context,
	container *tcpostgres.PostgresContainer,
	statement string,
) {
	t.Helper()
	command := []string{"psql", "-U", "entrypoint_test_admin", "-d", "postgres",
		"-v", "ON_ERROR_STOP=1", "-c", statement}
	exitCode, output, err := container.Exec(ctx, command)
	if err == nil && exitCode == 0 {
		return
	}
	body, _ := io.ReadAll(output)
	t.Fatalf("prepare schema executor role fixture: exit=%d err=%v output=%s",
		exitCode, err, body)
}
