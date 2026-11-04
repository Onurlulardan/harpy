package foundation

import (
	"strings"
	"testing"
)

func TestConfigRejectsUnsafeAndSharedSecrets(t *testing.T) {
	valid := Config{
		AdminDSN:       "postgres://admin:foundation-admin-test-secret@127.0.0.1:5432/postgres?sslmode=disable",
		Database:       "harpy_app",
		GoosePassword:  "goose-foundation-test-secret",
		APIPassword:    "api-foundation-test-secret01",
		WorkerPassword: "worker-foundation-test-secret",
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	tests := []Config{
		{AdminDSN: valid.AdminDSN, Database: "Harpy", GoosePassword: valid.GoosePassword, APIPassword: valid.APIPassword, WorkerPassword: valid.WorkerPassword},
		{AdminDSN: valid.AdminDSN, Database: "postgres", GoosePassword: valid.GoosePassword, APIPassword: valid.APIPassword, WorkerPassword: valid.WorkerPassword},
		{AdminDSN: valid.AdminDSN, Database: "harpy;drop", GoosePassword: valid.GoosePassword, APIPassword: valid.APIPassword, WorkerPassword: valid.WorkerPassword},
		{AdminDSN: "postgres://admin:foundation-admin-test-secret@127.0.0.1:5432/harpy_app?sslmode=disable", Database: "harpy_app", GoosePassword: valid.GoosePassword, APIPassword: valid.APIPassword, WorkerPassword: valid.WorkerPassword},
		{AdminDSN: valid.AdminDSN, Database: valid.Database, GoosePassword: "short", APIPassword: valid.APIPassword, WorkerPassword: valid.WorkerPassword},
		{AdminDSN: valid.AdminDSN, Database: valid.Database, GoosePassword: valid.GoosePassword, APIPassword: valid.GoosePassword, WorkerPassword: valid.WorkerPassword},
	}
	for _, cfg := range tests {
		if err := cfg.Validate(); err == nil {
			t.Fatalf("accepted invalid config %#v", cfg)
		}
	}
}

func TestDatabaseNameRejectsMaintenanceDatabases(t *testing.T) {
	for _, name := range []string{"postgres", "template0", "template1", "Bad", "a;b", ""} {
		if ValidDatabaseName(name) {
			t.Fatalf("accepted %q", name)
		}
	}
	if !ValidDatabaseName("harpy_app") {
		t.Fatal("rejected a safe database name")
	}
}

func TestRedactRemovesSecretsAndUserinfo(t *testing.T) {
	message := "dial postgres://admin:foundation-admin-test-secret@db/postgres failed goose-foundation-test-secret"
	got := Redact(message, []string{
		"postgres://admin:foundation-admin-test-secret@db/postgres",
		"foundation-admin-test-secret",
		"goose-foundation-test-secret",
	})
	if strings.Contains(got, "foundation-admin-test-secret") ||
		strings.Contains(got, "goose-foundation-test-secret") ||
		strings.Contains(got, "admin:") {
		t.Fatal(got)
	}
}
