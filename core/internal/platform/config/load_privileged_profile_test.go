package config

import (
	"strings"
	"testing"
	"time"
)

func TestPrivilegedProfilesLoadOnlyTheirCredential(t *testing.T) {
	bootstrap, err := loadClusterBootstrap(privilegedEnvironment(t, profileClusterBootstrap))
	if err != nil {
		t.Fatalf("bootstrap config: %v", err)
	}
	if bootstrap.Postgres().EndpointKind() != PostgresEndpointDirect ||
		bootstrap.Postgres().CleanupTimeout() != 5*time.Second ||
		bootstrap.Release().ReceiptPath().Value() == "" ||
		bootstrap.Artifacts().RolesManifestPath().Value() == "" {
		t.Fatalf("unexpected bootstrap config: %#v", bootstrap)
	}
	extension, err := loadExtensionAdmin(privilegedEnvironment(t, profileExtensionAdmin))
	if err != nil {
		t.Fatalf("extension config: %v", err)
	}
	if extension.Release().SigningKey().Value() != "ephemeral-test-key" {
		t.Fatal("extension signing key was not loaded")
	}
	if extension.Artifacts().ExtensionManifestPath().Value() == "" {
		t.Fatal("extension manifest was not loaded")
	}
}

func TestPrivilegedProfilesRejectOtherCredentialsAndRuntimeSurface(t *testing.T) {
	tests := []string{
		"HARPY_POSTGRES_DSN=postgresql://runtime:secret@db.invalid/app?sslmode=disable",
		"HARPY_EXTENSION_ADMIN_POSTGRES_DSN=postgresql://extension:secret@db.invalid/app?sslmode=disable",
		"HARPY_MANAGEMENT_LISTEN_ADDR=127.0.0.1:9000",
		"HARPY_OTLP_ENDPOINT=http://collector.invalid:4318",
	}
	for _, addition := range tests {
		entries := append(privilegedEnvironment(t, profileClusterBootstrap), addition)
		if _, err := loadClusterBootstrap(entries); err == nil || !strings.Contains(err.Error(), "unknown key") {
			t.Fatalf("expected profile isolation for %s, got %v", addition, err)
		}
	}
}

func TestPrivilegedProfilesRequireAbsoluteReleasePaths(t *testing.T) {
	entries := withEnvironment(
		privilegedEnvironment(t, profileClusterBootstrap),
		"HARPY_RELEASE_MANIFEST_PATH=relative.json",
	)
	if _, err := loadClusterBootstrap(entries); err == nil || !strings.Contains(err.Error(), "absolute") {
		t.Fatalf("expected absolute path failure, got %v", err)
	}
}

func TestPrivilegedProfilesRejectReceiptInputCollision(t *testing.T) {
	entries := privilegedEnvironment(t, profileClusterBootstrap)
	var manifest string
	for _, entry := range entries {
		if strings.HasPrefix(entry, "HARPY_RELEASE_MANIFEST_PATH=") {
			manifest = strings.TrimPrefix(entry, "HARPY_RELEASE_MANIFEST_PATH=")
		}
	}
	entries = withEnvironment(entries, "HARPY_BOOTSTRAP_RECEIPT_PATH="+manifest)
	if _, err := loadClusterBootstrap(entries); err == nil || !strings.Contains(err.Error(), "reuse") {
		t.Fatalf("expected receipt collision failure, got %v", err)
	}
}

func TestMigratorLoadsVerificationWithoutPrivilegedSigningKey(t *testing.T) {
	loaded, err := loadMigratorDeploy(migratorEnvironment(t))
	if err != nil {
		t.Fatalf("migrator verification config: %v", err)
	}
	if loaded.Verification().BootstrapReceiptPath().Value() == "" ||
		loaded.Verification().ExtensionReceiptPath().Value() == "" ||
		loaded.Verification().CatalogPackageSetPath().Value() == "" ||
		loaded.CatalogPackageKeyring().Validate() != nil ||
		loaded.Artifacts().MigrationManifestPath().Value() == "" {
		t.Fatal("migrator verification receipts were not loaded")
	}
	entries := append(migratorEnvironment(t), "HARPY_ATTESTATION_SIGNING_KEY=forbidden")
	if _, err := loadMigratorDeploy(entries); err == nil || !strings.Contains(err.Error(), "unknown key") {
		t.Fatalf("migrator accepted privileged signing key: %v", err)
	}
}

func TestPostgresArtifactPathCannotReuseReceipt(t *testing.T) {
	entries := privilegedEnvironment(t, profileClusterBootstrap)
	var receipt string
	for _, entry := range entries {
		if strings.HasPrefix(entry, "HARPY_BOOTSTRAP_RECEIPT_PATH=") {
			receipt = strings.TrimPrefix(entry, "HARPY_BOOTSTRAP_RECEIPT_PATH=")
		}
	}
	entries = withEnvironment(entries, "HARPY_ROLES_MANIFEST_PATH="+receipt)
	if _, err := loadClusterBootstrap(entries); err == nil || !strings.Contains(err.Error(), "distinct") {
		t.Fatalf("expected artifact path isolation failure, got %v", err)
	}
}
