package config

import (
	"strings"
	"testing"
)

func TestBootstrapRequiresExactMaintenanceDatabaseAndPrincipal(t *testing.T) {
	tests := []struct {
		name        string
		replacement string
		message     string
	}{
		{
			"maintenance database",
			"HARPY_BOOTSTRAP_MAINTENANCE_DSN=postgresql://harpy_cluster_admin:secret@db.invalid/app?sslmode=disable",
			"database postgres",
		},
		{
			"maintenance principal",
			"HARPY_BOOTSTRAP_MAINTENANCE_DSN=postgresql://postgres:secret@db.invalid/postgres?sslmode=disable",
			"harpy_cluster_admin",
		},
		{
			"target principal",
			"HARPY_BOOTSTRAP_TARGET_DSN=postgresql://harpy_owner:secret@db.invalid/app?sslmode=disable",
			"maintenance credential principal",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entries := withEnvironment(
				privilegedEnvironment(t, profileClusterBootstrap), test.replacement,
			)
			_, err := loadClusterBootstrap(entries)
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("expected %q failure, got %v", test.message, err)
			}
		})
	}
}

func TestPrivilegedProfilesRequireCanonicalLeaseID(t *testing.T) {
	entries := withEnvironment(
		privilegedEnvironment(t, profileExtensionAdmin),
		"HARPY_CREDENTIAL_LEASE_ID=55555555-5555-4555-8555-55555555555X",
	)
	if _, err := loadExtensionAdmin(entries); err == nil || !strings.Contains(err.Error(), "canonical") {
		t.Fatalf("expected canonical lease failure, got %v", err)
	}
}

func TestExtensionAdminRequiresExactPrincipal(t *testing.T) {
	entries := withEnvironment(
		privilegedEnvironment(t, profileExtensionAdmin),
		"HARPY_EXTENSION_ADMIN_POSTGRES_DSN=postgresql://postgres:secret@db.invalid/app?sslmode=disable",
	)
	if _, err := loadExtensionAdmin(entries); err == nil || !strings.Contains(err.Error(), "principal") {
		t.Fatalf("expected extension principal failure, got %v", err)
	}
}

func TestLegacySingleBootstrapDSNIsRejected(t *testing.T) {
	entries := append(
		privilegedEnvironment(t, profileClusterBootstrap),
		"HARPY_BOOTSTRAP_POSTGRES_DSN=postgresql://harpy_cluster_admin:secret@db.invalid/postgres?sslmode=disable",
	)
	if _, err := loadClusterBootstrap(entries); err == nil || !strings.Contains(err.Error(), "unknown key") {
		t.Fatalf("expected legacy key rejection, got %v", err)
	}
}
