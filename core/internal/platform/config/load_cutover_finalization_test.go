package config

import (
	"strings"
	"testing"
	"time"
)

func TestCutoverFinalizationLoadsExactOneShotProfile(t *testing.T) {
	loaded, err := loadCutoverFinalization(cutoverFinalizationEnvironment(t))
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Environment() != EnvironmentDevelopment || loaded.LogLevel() != LogInfo ||
		loaded.Postgres().ConnectTimeout() != 5*time.Second ||
		loaded.Artifacts().CredentialCleanupPath().Value() == "" ||
		loaded.Artifacts().FinalizerLeaseManifestPath().Value() == "" ||
		loaded.Artifacts().TerminalEvidencePath().Value() == "" {
		t.Fatal("finalization profile did not preserve exact values")
	}
}

func TestCutoverFinalizationRejectsWrongPrincipalAndForeignAuthority(t *testing.T) {
	wrong := withEnvironment(
		cutoverFinalizationEnvironment(t),
		"HARPY_CUTOVER_FINALIZER_DSN=postgresql://harpy_adopter:must-not-leak@target.invalid/app?sslmode=disable",
	)
	if _, err := loadCutoverFinalization(wrong); err == nil ||
		!strings.Contains(err.Error(), "unexpected database principal") ||
		strings.Contains(err.Error(), "must-not-leak") {
		t.Fatalf("expected redacted principal rejection, got %v", err)
	}
	for _, addition := range []string{
		"HARPY_CUTOVER_TARGET_DSN=postgresql://harpy_adopter:secret@target.invalid/app?sslmode=disable",
		"HARPY_ATTESTATION_SIGNING_KEY=forbidden",
		"HARPY_MANAGEMENT_LISTEN_ADDR=127.0.0.1:9000",
	} {
		if _, err := loadCutoverFinalization(
			append(cutoverFinalizationEnvironment(t), addition),
		); err == nil || !strings.Contains(err.Error(), "unknown key") {
			t.Fatalf("finalization profile accepted foreign key: %v", err)
		}
	}
}

func TestCutoverFinalizationRequiresAllDistinctEvidence(t *testing.T) {
	missing := removeEnvironmentKey(
		cutoverFinalizationEnvironment(t),
		"HARPY_CUTOVER_CREDENTIAL_CLEANUP_RECEIPT_PATH",
	)
	if _, err := loadCutoverFinalization(missing); err == nil {
		t.Fatal("finalization accepted missing cleanup evidence")
	}
	entries := cutoverFinalizationEnvironment(t)
	var trustPath string
	for _, entry := range entries {
		if strings.HasPrefix(entry, "HARPY_CUTOVER_TRUST_ROOT_PATH=") {
			trustPath = strings.TrimPrefix(entry, "HARPY_CUTOVER_TRUST_ROOT_PATH=")
		}
	}
	collision := withEnvironment(
		entries, "HARPY_CUTOVER_FINALIZER_LEASE_SIGNATURE_PATH="+trustPath,
	)
	if _, err := loadCutoverFinalization(collision); err == nil ||
		!strings.Contains(err.Error(), "distinct") {
		t.Fatalf("expected finalization artifact collision rejection, got %v", err)
	}
}
