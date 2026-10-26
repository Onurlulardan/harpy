package config

import (
	"strings"
	"testing"
)

func TestProfilesRejectUnknownHarpyKeys(t *testing.T) {
	tests := []struct {
		name string
		load func([]string) error
		key  string
	}{
		{"api unknown", func(entries []string) error { _, err := loadAPI(entries); return err }, "HARPY_TYPO=value"},
		{"worker public", func(entries []string) error { _, err := loadWorker(entries); return err }, "HARPY_PUBLIC_LISTEN_ADDR=127.0.0.1:8000"},
		{"deploy management", func(entries []string) error { _, err := loadMigratorDeploy(entries); return err }, "HARPY_MANAGEMENT_LISTEN_ADDR=127.0.0.1:9000"},
		{"wrong case", func(entries []string) error { _, err := loadAPI(entries); return err }, "harpy_log_level=debug"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.load(append(validEnvironment(), test.key))
			if err == nil || !strings.Contains(err.Error(), "unknown key") {
				t.Fatalf("expected unknown-key error, got %v", err)
			}
		})
	}
}

func TestProfilesAcceptOnlyTheirLongRunningSurface(t *testing.T) {
	if _, err := loadWorker(validWorkerEnvironment()); err != nil {
		t.Fatalf("worker defaults rejected: %v", err)
	}
	if _, err := loadSchemaExecutor(validEnvironment()); err != nil {
		t.Fatalf("schema executor defaults rejected: %v", err)
	}
	if _, err := loadMigratorDeploy(migratorEnvironment(t)); err != nil {
		t.Fatalf("migrator deploy defaults rejected: %v", err)
	}
}

func TestRequiredEnvironmentAndSecretAreFailClosed(t *testing.T) {
	for _, entries := range [][]string{
		{"HARPY_POSTGRES_DSN=postgresql://example.invalid/db"},
		{"HARPY_ENVIRONMENT=development"},
		{"HARPY_ENVIRONMENT=development", "HARPY_POSTGRES_DSN="},
	} {
		if _, err := loadAPI(entries); err == nil {
			t.Fatalf("expected required config failure for %#v", entries)
		}
	}
}

func TestPostgresCleanupTimeoutRequiresExactFiveSeconds(t *testing.T) {
	entries := append(validEnvironment(), "HARPY_POSTGRES_CLEANUP_TIMEOUT=4s")
	if _, err := loadSchemaExecutor(entries); err == nil {
		t.Fatal("non-contract cleanup timeout accepted")
	}
}

func TestTelemetryFlushMustFitInsideDrainDeadline(t *testing.T) {
	entries := withEnvironment(
		validWorkerEnvironment(),
		"HARPY_DRAIN_TIMEOUT=5s", "HARPY_OTLP_EXPORT_TIMEOUT=6s",
	)
	if _, err := loadWorker(entries); err == nil {
		t.Fatal("telemetry timeout exceeding drain deadline accepted")
	}
}
