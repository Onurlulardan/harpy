package main

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRevisionTwelveMovesRuntimeAndBackupOntoHarpyAPI(t *testing.T) {
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	output := writeRevisionTwelveProfileSet(t, now)
	apiBefore, err := os.ReadFile(filepath.Join(output, "postgres-api-password.txt"))
	if err != nil {
		t.Fatal(err)
	}
	restoreBefore, err := os.ReadFile(filepath.Join(output, "postgres-restore-dsn.txt"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := ensureArtifactSet(output, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Generated ||
		result.Manifest.ArtifactContractRevision != currentArtifactContractRevision {
		t.Fatalf("upgrade result = %#v", result.Manifest)
	}
	apiAfter, err := os.ReadFile(filepath.Join(output, "postgres-api-password.txt"))
	if err != nil || string(apiAfter) != string(apiBefore) {
		t.Fatal("foundation profile upgrade rotated the api password")
	}
	assertPrincipalDSN(t, output, "postgres-runtime-dsn.txt", "harpy_api", string(apiAfter))
	assertPrincipalDSN(t, output, "postgres-backup-dsn.txt", "harpy_api", string(apiAfter))
	restoreAfter, err := os.ReadFile(filepath.Join(output, "postgres-restore-dsn.txt"))
	if err != nil || string(restoreAfter) != string(restoreBefore) {
		t.Fatal("foundation profile upgrade changed the restore DSN")
	}
}

func writeRevisionTwelveProfileSet(t *testing.T, now time.Time) string {
	t.Helper()
	set, err := generateArtifactSet(now)
	if err != nil {
		t.Fatal(err)
	}
	runtimePassword := string(set.files["postgres-runtime-password.txt"])
	backupPassword := string(set.files["postgres-backup-password.txt"])
	set.files["postgres-runtime-dsn.txt"] = []byte(postgresDSN(
		"harpy_v2_runtime", runtimePassword, "harpy"))
	set.files["postgres-backup-dsn.txt"] = []byte(postgresDSN(
		"harpy_backup", backupPassword, "harpy"))
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		t.Fatal(err)
	}
	manifest.ArtifactContractRevision = foundationProfileUpgradeSourceRevision
	output := filepath.Join(t.TempDir(), "secrets")
	if err := os.Mkdir(output, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, file := range manifest.Files {
		if err := os.WriteFile(
			filepath.Join(output, file.Name), set.files[file.Name], 0o600,
		); err != nil {
			t.Fatal(err)
		}
	}
	content, err := marshalJSON(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(output, manifestFileName), content, 0o600); err != nil {
		t.Fatal(err)
	}
	return output
}

func assertPrincipalDSN(t *testing.T, output, name, username, password string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(output, name))
	if err != nil {
		t.Fatal(err)
	}
	value, err := url.Parse(string(content))
	actual, exists := value.User.Password()
	if err != nil || value.User.Username() != username || !exists || actual != password {
		t.Fatalf("%s principal = %s", name, value.User.Username())
	}
}
