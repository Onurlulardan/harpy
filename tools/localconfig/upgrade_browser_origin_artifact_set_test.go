package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRevisionEightAddsLocalBrowserOriginWithoutRotatingSecrets(t *testing.T) {
	now := time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
	set, err := generateArtifactSet(now)
	if err != nil {
		t.Fatal(err)
	}
	set.files["api.env"] = bytes.Replace(
		set.files["api.env"], currentLocalBrowserOrigins, legacyLocalBrowserOrigins, 1,
	)
	output := publishRevisionEightSet(t, now, set)
	secretBefore := readUpgradeSecret(t, output)
	result, err := ensureArtifactSet(output, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	secretAfter := readUpgradeSecret(t, output)
	apiEnvironment, readErr := os.ReadFile(filepath.Join(output, "api.env"))
	if readErr != nil || !bytes.Contains(apiEnvironment, currentLocalBrowserOrigins) {
		t.Fatal("local browser origin upgrade did not refresh api.env")
	}
	if !bytes.Equal(secretBefore, secretAfter) {
		t.Fatal("local browser origin upgrade rotated the runtime database secret")
	}
	assertBrowserOriginUpgradeResult(t, result)
}

func TestRevisionEightKeepsForgeGuardOriginWithoutRotatingSecrets(t *testing.T) {
	now := time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
	options := forgeGuardOptionsForTest(t, "connected")
	set, err := generateArtifactSetFor(now, options)
	if err != nil {
		t.Fatal(err)
	}
	output := publishRevisionEightSet(t, now, set)
	secretBefore := readUpgradeSecret(t, output)
	result, err := ensureForgeGuardArtifactSet(
		output, now.Add(time.Minute), options,
	)
	if err != nil {
		t.Fatal(err)
	}
	secretAfter := readUpgradeSecret(t, output)
	if !bytes.Equal(secretBefore, secretAfter) {
		t.Fatal("ForgeGuard browser origin upgrade rotated the runtime database secret")
	}
	assertBrowserOriginUpgradeResult(t, result)
}

func publishRevisionEightSet(
	t *testing.T,
	now time.Time,
	set *artifactSet,
) string {
	t.Helper()
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		t.Fatal(err)
	}
	manifest.ArtifactContractRevision = browserOriginUpgradeSourceRevision
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
	manifestContent, err := marshalJSON(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err = writeArtifactFile(
		filepath.Join(output, manifestFileName), manifestContent,
	); err != nil {
		t.Fatal(err)
	}
	return output
}

func readUpgradeSecret(t *testing.T, output string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(output, "postgres-runtime-password.txt"))
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func assertBrowserOriginUpgradeResult(t *testing.T, result ensureResult) {
	t.Helper()
	if !result.Generated ||
		result.Manifest.ArtifactContractRevision != currentArtifactContractRevision {
		t.Fatalf("browser origin upgrade result = %#v", result)
	}
}
