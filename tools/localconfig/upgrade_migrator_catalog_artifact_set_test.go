package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRevisionTenRestoresMigratorCatalogEnvironment(t *testing.T) {
	now := time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC)
	set, err := generateArtifactSet(now)
	if err != nil {
		t.Fatal(err)
	}
	set.files["migrator.env"] = withoutMigratorCatalogKeys(set.files["migrator.env"])
	output := publishMigratorCatalogSourceSet(t, now, set)
	secretBefore := readUpgradeSecret(t, output)
	result, err := ensureArtifactSet(output, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	environment, err := os.ReadFile(filepath.Join(output, "migrator.env"))
	if err != nil {
		t.Fatal(err)
	}
	for key, value := range migratorCatalogPackageEnvironment() {
		if !bytes.Contains(environment, []byte(key+"="+value+"\n")) {
			t.Fatalf("migrator catalog upgrade did not restore %s", key)
		}
	}
	if !bytes.Equal(secretBefore, readUpgradeSecret(t, output)) {
		t.Fatal("migrator catalog upgrade rotated the runtime database secret")
	}
	if !result.Generated ||
		result.Manifest.ArtifactContractRevision != currentArtifactContractRevision {
		t.Fatalf("migrator catalog upgrade result = %#v", result)
	}
}

func withoutMigratorCatalogKeys(document []byte) []byte {
	removed := migratorCatalogPackageEnvironment()
	kept := make([]string, 0)
	for _, line := range strings.Split(strings.TrimSpace(string(document)), "\n") {
		key, _, _ := strings.Cut(line, "=")
		if _, drop := removed[key]; drop {
			continue
		}
		kept = append(kept, line)
	}
	return []byte(strings.Join(kept, "\n") + "\n")
}

func publishMigratorCatalogSourceSet(
	t *testing.T,
	now time.Time,
	set *artifactSet,
) string {
	t.Helper()
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		t.Fatal(err)
	}
	manifest.ArtifactContractRevision = migratorCatalogUpgradeSourceRevision
	output := filepath.Join(t.TempDir(), "secrets")
	if err = os.Mkdir(output, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, file := range manifest.Files {
		if err = writeArtifactFile(
			filepath.Join(output, file.Name), set.files[file.Name],
		); err != nil {
			t.Fatal(err)
		}
	}
	content, err := marshalJSON(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeArtifactFile(
		filepath.Join(output, manifestFileName), content,
	); err != nil {
		t.Fatal(err)
	}
	return output
}
