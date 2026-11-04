package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/google/uuid"
)

func TestPublishDevelopmentBootstrapAuthorityRejectsInvalidStoredIdentity(t *testing.T) {
	now := time.Date(2026, 11, 4, 16, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		mutate func(*testing.T, string)
	}{
		{name: "corrupt", mutate: func(t *testing.T, path string) {
			if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "schema invalid", mutate: func(t *testing.T, path string) {
			content := []byte(`{"contract_version":"bootstrap-authority-manifest.v1"}`)
			if err := os.WriteFile(path, content, 0o600); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := filepath.Join(t.TempDir(), "current")
			if err := publishDevelopmentBootstrapAuthority(directory, now); err != nil {
				t.Fatal(err)
			}
			previous := readPublishedAuthority(t, directory)
			test.mutate(t, filepath.Join(directory, "bootstrap-authority.json"))
			if err := publishDevelopmentBootstrapAuthority(directory, now); err != nil {
				t.Fatal(err)
			}
			next := readPublishedAuthority(t, directory)
			if next.EnvironmentID == previous.EnvironmentID ||
				next.ReleaseID == previous.ReleaseID ||
				next.ManifestID == previous.ManifestID {
				t.Fatal("invalid stored authority reused its identity")
			}
			requireUUIDv7(t, next.EnvironmentID, next.ReleaseID, next.ManifestID)
		})
	}
}

func assertPublishedLifetime(
	t *testing.T, manifest runtimev1.BootstrapAuthorityManifest, now time.Time,
) {
	t.Helper()
	issued := now.Add(-time.Minute).Format(time.RFC3339)
	authorityExpires := now.Add(24 * time.Hour).Format(time.RFC3339)
	leaseExpires := now.Add(-time.Minute).Add(15 * time.Minute).Format(time.RFC3339)
	if manifest.IssuedAt != issued || manifest.ExpiresAt != authorityExpires {
		t.Fatalf("authority window = %s %s", manifest.IssuedAt, manifest.ExpiresAt)
	}
	principals := []string{"harpy_cluster_admin", "harpy_extension_admin"}
	if len(manifest.CredentialLeases) != len(principals) {
		t.Fatalf("lease count = %d", len(manifest.CredentialLeases))
	}
	for index, lease := range manifest.CredentialLeases {
		if lease.Principal != principals[index] || lease.NotBefore != issued ||
			lease.ExpiresAt != leaseExpires || lease.MaxDurationSeconds != 900 {
			t.Fatalf("lease %d window = %+v", index, lease)
		}
	}
}

func assertPublishedOmitsRetiredRoles(t *testing.T, directory string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(directory, "bootstrap-authority.json"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	for _, name := range []string{
		"harpy_v2_runtime", "harpy_cutover_authority_owner", "harpy_cutover_decoder",
		"harpy_cutover_finalizer", "harpy_adopter", "harpy_v1_runtime", "harpy_runtime",
		"harpy_backup", "harpy_restore", "harpy_reactivation_observer",
		"harpy_legacy_extension_owner", "harpy_backup_reader",
	} {
		if strings.Contains(text, name) {
			t.Fatalf("published authority contains %s", name)
		}
	}
}

func requireUUIDv7(t *testing.T, values ...string) {
	t.Helper()
	for _, value := range values {
		parsed, err := uuid.Parse(value)
		if err != nil || parsed.Version() != uuid.Version(7) {
			t.Fatalf("identity %s is not UUIDv7", value)
		}
	}
}
