package main

import "testing"

const validAuthorityTestDSN = "postgres://harpy_cluster_admin:secret@127.0.0.1:5432/harpy?sslmode=disable"

func TestAuthorityDatabaseConfigUsesExplicitURLOnly(t *testing.T) {
	clearPostgresEnvironment(t)
	config, err := parseAuthorityDatabaseConfig(validAuthorityTestDSN, "test", true)
	if err != nil {
		t.Fatalf("parse explicit authority DSN: %v", err)
	}
	if config.User != "harpy_cluster_admin" || config.Host != "127.0.0.1" ||
		config.Database != "harpy" || config.Password != "secret" ||
		config.RuntimeParams["application_name"] != authorityApplicationName {
		t.Fatalf("explicit authority DSN changed: %+v", config.Config)
	}
}

func TestAuthorityDatabaseConfigRejectsAmbientPostgresDefaults(t *testing.T) {
	for _, name := range authorityPostgresEnvironment {
		t.Run(name, func(t *testing.T) {
			clearPostgresEnvironment(t)
			t.Setenv(name, "ambient-value")
			if _, err := parseAuthorityDatabaseConfig(
				validAuthorityTestDSN, "test", true,
			); err == nil {
				t.Fatalf("accepted ambient %s", name)
			}
		})
	}
}

func TestAuthorityDatabaseConfigRequiresExplicitIdentity(t *testing.T) {
	clearPostgresEnvironment(t)
	for name, dsn := range map[string]string{
		"URL form": "host=127.0.0.1 user=role dbname=harpy password=secret",
		"user":     "postgres://:secret@127.0.0.1:5432/harpy?sslmode=disable",
		"host":     "postgres://role:secret@/harpy?sslmode=disable",
		"database": "postgres://role:secret@127.0.0.1:5432?sslmode=disable",
		"password": "postgres://role@127.0.0.1:5432/harpy?sslmode=disable",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseAuthorityDatabaseConfig(dsn, "test", true); err == nil {
				t.Fatalf("accepted DSN without explicit %s", name)
			}
		})
	}
}

func TestAuthorityDatabaseConfigRejectsRedirectingKeys(t *testing.T) {
	clearPostgresEnvironment(t)
	for _, key := range []string{"service=redirect", "servicefile=redirect",
		"passfile=redirect", "host=redirect", "user=redirect", "dbname=redirect"} {
		dsn := validAuthorityTestDSN + "&" + key
		if _, err := parseAuthorityDatabaseConfig(dsn, "test", true); err == nil {
			t.Fatalf("accepted redirecting key %s", key)
		}
	}
}

func TestAuthorityDatabaseConfigDisablesImplicitPGPass(t *testing.T) {
	clearPostgresEnvironment(t)
	dsn := "postgres://harpy_cluster_admin@127.0.0.1:5432/harpy?sslmode=disable"
	config, err := parseAuthorityDatabaseConfig(dsn, "test", false)
	if err != nil || config.Password != "" {
		t.Fatalf("passwordless explicit profile inherited a password: config=%+v err=%v",
			config, err)
	}
}

func clearPostgresEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range authorityPostgresEnvironment {
		t.Setenv(name, "")
	}
}
