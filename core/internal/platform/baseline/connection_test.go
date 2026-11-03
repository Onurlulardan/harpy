package baseline

import "testing"

func TestOwnerSessionPinsDatabaseOwnerWithoutConnecting(t *testing.T) {
	const dsn = "postgres://harpy_migrator@127.0.0.1:1/harpy?sslmode=disable"
	owner, err := sessionConfig(dsn, "platform, pg_catalog", true)
	if err != nil {
		t.Fatalf("parse owner session: %v", err)
	}
	if owner.RuntimeParams["role"] != databaseOwnerRole ||
		owner.RuntimeParams["search_path"] != "platform, pg_catalog" ||
		owner.User != "harpy_migrator" || owner.Database != "harpy" {
		t.Fatalf("owner session = user %q database %q params %#v",
			owner.User, owner.Database, owner.RuntimeParams)
	}
	login, err := sessionConfig(dsn, "pg_catalog", false)
	if err != nil {
		t.Fatalf("parse login session: %v", err)
	}
	if _, pinned := login.RuntimeParams["role"]; pinned ||
		login.RuntimeParams["search_path"] != "pg_catalog" {
		t.Fatalf("login session params = %#v", login.RuntimeParams)
	}
}

func TestSessionConfigRejectsInvalidDSN(t *testing.T) {
	if _, err := sessionConfig("://", "", true); err == nil {
		t.Fatal("invalid DSN was accepted")
	}
}
