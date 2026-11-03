package baseline

import "testing"

func TestMigratorSessionAssumesDatabaseOwnerWithoutConnecting(t *testing.T) {
	const dsn = "postgres://harpy_migrator@127.0.0.1:1/harpy?sslmode=disable"
	migrator, err := sessionConfig(dsn, "platform, pg_catalog")
	if err != nil {
		t.Fatalf("parse migrator session: %v", err)
	}
	if migrator.RuntimeParams["role"] != databaseOwnerRole ||
		migrator.RuntimeParams["search_path"] != "platform, pg_catalog" ||
		migrator.User != migratorLoginRole || migrator.Database != "harpy" {
		t.Fatalf("migrator session = user %q database %q params %#v",
			migrator.User, migrator.Database, migrator.RuntimeParams)
	}
}

func TestDatabaseOwnerSessionKeepsItsLoginRole(t *testing.T) {
	const dsn = "postgresql://postgres:postgres@127.0.0.1:1/harpy_test?sslmode=disable"
	owner, err := sessionConfig(dsn, "pg_catalog")
	if err != nil {
		t.Fatalf("parse owner session: %v", err)
	}
	if _, pinned := owner.RuntimeParams["role"]; pinned ||
		owner.User != "postgres" || owner.Database != "harpy_test" ||
		owner.RuntimeParams["search_path"] != "pg_catalog" {
		t.Fatalf("owner session = user %q database %q params %#v",
			owner.User, owner.Database, owner.RuntimeParams)
	}
}

func TestSessionConfigRejectsInvalidDSN(t *testing.T) {
	if _, err := sessionConfig("://", ""); err == nil {
		t.Fatal("invalid DSN was accepted")
	}
}
