package config

import (
	"strings"
	"testing"
	"time"
)

func TestCutoverAdoptionLoadsExactOneShotProfile(t *testing.T) {
	loaded, err := loadCutoverAdoption(cutoverEnvironment(t))
	if err != nil {
		t.Fatalf("cutover config: %v", err)
	}
	if loaded.Environment() != EnvironmentDevelopment || loaded.LogLevel() != LogInfo {
		t.Fatalf("unexpected cutover runtime values: %#v", loaded)
	}
	if loaded.SourcePostgres().ConnectTimeout() != 5*time.Second ||
		loaded.TargetPostgres().CleanupTimeout() != 5*time.Second {
		t.Fatal("cutover PostgreSQL timeout defaults drifted")
	}
	if loaded.Artifacts().SessionSealReceiptPath().Value() == "" ||
		loaded.Artifacts().BatchPath().Value() == "" {
		t.Fatal("cutover evidence paths were not loaded")
	}
	if loaded.Telemetry().Enabled() {
		t.Fatal("development telemetry should default to no-export")
	}
}

func TestCutoverAdoptionRequiresExactDatabasePrincipals(t *testing.T) {
	tests := []struct {
		key string
		dsn string
	}{
		{"HARPY_CUTOVER_SOURCE_DSN", "postgresql://harpy_adopter:must-not-leak@source.invalid/app?sslmode=disable"},
		{"HARPY_CUTOVER_TARGET_DSN", "postgresql://harpy_cutover_decoder:must-not-leak@target.invalid/app?sslmode=disable"},
	}
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			_, err := loadCutoverAdoption(withEnvironment(cutoverEnvironment(t), test.key+"="+test.dsn))
			if err == nil || !strings.Contains(err.Error(), "unexpected database principal") {
				t.Fatalf("expected principal rejection, got %v", err)
			}
			if strings.Contains(err.Error(), "must-not-leak") || strings.Contains(err.Error(), ".invalid") {
				t.Fatalf("cutover DSN leaked in error: %v", err)
			}
		})
	}
}

func TestCutoverAdoptionRequiresProductionTLSAndTelemetry(t *testing.T) {
	entries := withEnvironment(
		cutoverEnvironment(t),
		"HARPY_ENVIRONMENT=production",
		"HARPY_CUTOVER_SOURCE_DSN=postgresql://harpy_cutover_decoder:secret@source.invalid/app?sslmode=verify-full",
		"HARPY_CUTOVER_TARGET_DSN=postgresql://harpy_adopter:secret@target.invalid/app?sslmode=verify-full",
	)
	if _, err := loadCutoverAdoption(entries); err == nil || !strings.Contains(err.Error(), "OTLP_ENDPOINT") {
		t.Fatalf("expected production OTLP requirement, got %v", err)
	}
	entries = append(entries, "HARPY_OTLP_ENDPOINT=https://collector.invalid:4318")
	if _, err := loadCutoverAdoption(entries); err != nil {
		t.Fatalf("secure production cutover profile rejected: %v", err)
	}
	weakTLS := withEnvironment(entries,
		"HARPY_CUTOVER_SOURCE_DSN=postgresql://harpy_cutover_decoder:secret@source.invalid/app?sslmode=disable",
	)
	if _, err := loadCutoverAdoption(weakTLS); err == nil {
		t.Fatal("production cutover profile accepted disabled source TLS")
	}
}

func TestCutoverAdoptionRejectsForeignCredentialsAndListeners(t *testing.T) {
	for _, addition := range []string{
		"HARPY_POSTGRES_DSN=postgresql://harpy_migrator:secret@db.invalid/app?sslmode=disable",
		"HARPY_ATTESTATION_SIGNING_KEY=forbidden",
		"HARPY_RELEASE_MANIFEST_PATH=/release.json",
		"HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH=/authority.json",
		"HARPY_MANAGEMENT_LISTEN_ADDR=127.0.0.1:9000",
		"HARPY_PUBLIC_LISTEN_ADDR=127.0.0.1:8000",
	} {
		_, err := loadCutoverAdoption(append(cutoverEnvironment(t), addition))
		if err == nil || !strings.Contains(err.Error(), "unknown key") {
			t.Fatalf("cutover profile accepted %s: %v", strings.SplitN(addition, "=", 2)[0], err)
		}
	}
}

func TestCutoverAdoptionRejectsMissingOrCollidingEvidence(t *testing.T) {
	missing := removeEnvironmentKey(cutoverEnvironment(t), "HARPY_CUTOVER_SESSION_SEAL_RECEIPT_PATH")
	if _, err := loadCutoverAdoption(missing); err == nil {
		t.Fatal("cutover profile accepted missing session seal receipt")
	}
	entries := cutoverEnvironment(t)
	var leasePath string
	for _, entry := range entries {
		if strings.HasPrefix(entry, "HARPY_CUTOVER_LEASE_MANIFEST_PATH=") {
			leasePath = strings.TrimPrefix(entry, "HARPY_CUTOVER_LEASE_MANIFEST_PATH=")
		}
	}
	collision := withEnvironment(entries, "HARPY_CUTOVER_LEASE_SIGNATURE_PATH="+leasePath)
	if _, err := loadCutoverAdoption(collision); err == nil || !strings.Contains(err.Error(), "distinct") {
		t.Fatalf("expected artifact collision rejection, got %v", err)
	}
}

func TestCutoverAdoptionBoundsPostgresTimeouts(t *testing.T) {
	for _, timeout := range []string{
		"HARPY_POSTGRES_CONNECT_TIMEOUT=31s",
		"HARPY_POSTGRES_CLEANUP_TIMEOUT=500ms",
	} {
		if _, err := loadCutoverAdoption(withEnvironment(cutoverEnvironment(t), timeout)); err == nil {
			t.Fatalf("cutover profile accepted unbounded timeout %s", timeout)
		}
	}
}
