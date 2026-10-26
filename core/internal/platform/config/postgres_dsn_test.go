package config

import (
	"strings"
	"testing"
	"time"
)

func TestPostgresRequiresDirectEndpointAndEnvironmentTLSMode(t *testing.T) {
	tests := []struct {
		name         string
		replacements []string
	}{
		{"missing endpoint kind", []string{"HARPY_POSTGRES_ENDPOINT_KIND"}},
		{"proxy endpoint kind", []string{"HARPY_POSTGRES_ENDPOINT_KIND=pgbouncer"}},
		{"missing sslmode", []string{"HARPY_POSTGRES_DSN=postgresql://user@example.invalid/db"}},
		{"development weak TLS", []string{"HARPY_POSTGRES_DSN=postgresql://user@example.invalid/db?sslmode=require"}},
		{"unsupported parameter", []string{"HARPY_POSTGRES_DSN=postgresql://user@example.invalid/db?sslmode=disable&role=owner"}},
		{"multiple hosts", []string{"HARPY_POSTGRES_DSN=postgresql://user@one.invalid,two.invalid/db?sslmode=disable"}},
		{"disabled TLS files", []string{"HARPY_POSTGRES_DSN=postgresql://user@example.invalid/db?sslmode=disable&sslrootcert=/ca.pem"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entries := validAPIEnvironment()
			for _, replacement := range test.replacements {
				if !strings.Contains(replacement, "=") {
					entries = removeEnvironmentKey(entries, replacement)
					continue
				}
				entries = withEnvironment(entries, replacement)
			}
			if _, err := loadAPI(entries); err == nil {
				t.Fatal("expected PostgreSQL contract rejection")
			}
		})
	}
}

func TestPostgresAllowsVerifyFullInDevelopment(t *testing.T) {
	entries := withEnvironment(validAPIEnvironment(),
		"HARPY_POSTGRES_DSN=postgresql://user@example.invalid/db?sslmode=verify-full",
	)
	if _, err := loadAPI(entries); err != nil {
		t.Fatalf("development verify-full rejected: %v", err)
	}
}

func TestPostgresAcceptsProductionVerifyFullWithoutLeakingDSN(t *testing.T) {
	entries := withEnvironment(validAPIEnvironment(),
		"HARPY_ENVIRONMENT=production",
		"HARPY_POSTGRES_DSN=postgresql://user:must-not-leak@db.invalid/app?sslmode=verify-full",
		"HARPY_OTLP_ENDPOINT=https://collector.invalid:4318",
	)
	loaded, err := loadAPI(entries)
	if err != nil {
		t.Fatalf("production DSN rejected: %v", err)
	}
	postgres := loaded.LongRunning().Runtime().Postgres()
	if postgres.EndpointKind() != PostgresEndpointDirect || postgres.CleanupTimeout() != 5*time.Second {
		t.Fatalf("unexpected PostgreSQL contract: %#v", postgres)
	}
	invalid := withEnvironment(entries,
		"HARPY_POSTGRES_DSN=postgresql://user:must-not-leak@db.invalid/app?sslmode=disable",
	)
	_, err = loadAPI(invalid)
	if err == nil || strings.Contains(err.Error(), "must-not-leak") || strings.Contains(err.Error(), "db.invalid") {
		t.Fatalf("DSN error leaked secret material: %v", err)
	}
}

func removeEnvironmentKey(entries []string, key string) []string {
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !strings.HasPrefix(entry, key+"=") {
			result = append(result, entry)
		}
	}
	return result
}
