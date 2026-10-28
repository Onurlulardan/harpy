package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCurrentContractRevisionIsSerialized(t *testing.T) {
	set, now := generatedSetForTest(t)
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		t.Fatal(err)
	}
	content, err := marshalJSON(manifest)
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf(`"artifact_contract_revision":%d`, currentArtifactContractRevision)
	if !bytes.Contains(content, []byte(want)) {
		t.Fatalf("current contract revision is absent: %s", content)
	}
}

func TestPriorContractRevisionIsNotCurrentButExplicitlyResettable(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	output := contractRevisionSetForTest(t, now, 0)
	if _, err := ensureArtifactSet(output, now.Add(time.Hour)); err == nil {
		t.Fatal("normal ensure accepted prior artifact contract revision")
	}
	removed, _, err := resetArtifactSet(output, now.Add(time.Hour))
	if err != nil || !removed {
		t.Fatalf("prior revision reset: removed=%t err=%v", removed, err)
	}
}

func TestDamagedPriorContractRevisionRefusesReset(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	output := contractRevisionSetForTest(t, now, 0)
	path := filepath.Join(output, "rabbitmq.conf")
	if err := os.WriteFile(path, []byte("damaged"), 0o600); err != nil {
		t.Fatal(err)
	}
	removed, _, err := resetArtifactSet(output, now.Add(time.Hour))
	if err == nil || removed {
		t.Fatalf("damaged prior revision reset: removed=%t err=%v", removed, err)
	}
	if _, statErr := os.Lstat(output); statErr != nil {
		t.Fatalf("damaged prior revision was removed: %v", statErr)
	}
}

func TestUnknownFutureContractRevisionRefusesReset(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	output := contractRevisionSetForTest(t, now, currentArtifactContractRevision+1)
	removed, _, err := resetArtifactSet(output, now.Add(time.Hour))
	if err == nil || removed {
		t.Fatalf("future revision reset: removed=%t err=%v", removed, err)
	}
}

func contractRevisionSetForTest(t *testing.T, now time.Time, revision int) string {
	t.Helper()
	output := filepath.Join(t.TempDir(), "secrets")
	if _, err := ensureArtifactSet(output, now); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(output, manifestFileName)
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := decodeSetManifest(content)
	if err != nil {
		t.Fatal(err)
	}
	manifest.ArtifactContractRevision = revision
	content, err = marshalJSON(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if revision == 0 && bytes.Contains(content, []byte("artifact_contract_revision")) {
		t.Fatal("revision zero must remain compatible with field-absent manifests")
	}
	if err := os.WriteFile(manifestPath, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return output
}
