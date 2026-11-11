package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRevisionFourteenAddsDocsOriginWithoutRotatingSecrets(t *testing.T) {
	now := time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
	set, err := generateArtifactSet(now)
	if err != nil {
		t.Fatal(err)
	}
	set.files["api.env"] = bytes.Replace(
		set.files["api.env"], currentLocalBrowserOrigins, previousLocalBrowserOrigins, 1,
	)
	output := publishRevisionSet(t, now, set, docsOriginUpgradeSourceRevision)
	secretBefore := readUpgradeSecret(t, output)
	result, err := ensureArtifactSet(output, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	secretAfter := readUpgradeSecret(t, output)
	apiEnvironment, readErr := os.ReadFile(filepath.Join(output, "api.env"))
	if readErr != nil || !bytes.Contains(apiEnvironment, currentLocalBrowserOrigins) {
		t.Fatal("docs origin upgrade did not refresh api.env")
	}
	if bytes.Contains(apiEnvironment, previousLocalBrowserOrigins) {
		t.Fatal("docs origin upgrade left the previous origin list")
	}
	if !bytes.Equal(secretBefore, secretAfter) {
		t.Fatal("docs origin upgrade rotated the runtime database secret")
	}
	if !result.Generated ||
		result.Manifest.ArtifactContractRevision != currentArtifactContractRevision {
		t.Fatalf("docs origin upgrade result = %#v", result)
	}
}

func publishRevisionSet(
	t *testing.T,
	now time.Time,
	set *artifactSet,
	revision int,
) string {
	t.Helper()
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		t.Fatal(err)
	}
	manifest.ArtifactContractRevision = revision
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
