package deployevidence

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadArtifactAcceptsBoundedRegularFile(t *testing.T) {
	path := writeArtifactFixture(t, []byte("{}"))
	document, err := readArtifact(path, 2, "deployment")
	if err != nil {
		t.Fatalf("read bounded artifact: %v", err)
	}
	if string(document.bytes) != "{}" || document.resolvedPath != path {
		t.Fatalf("unexpected artifact document: %+v", document)
	}
}

func TestReadArtifactRejectsOversizeWithoutLeakingPath(t *testing.T) {
	path := writeArtifactFixture(t, []byte("oversize"))
	_, err := readArtifact(path, 2, "deployment")
	if err == nil {
		t.Fatal("oversize artifact accepted")
	}
	if strings.Contains(err.Error(), path) {
		t.Fatalf("artifact path leaked: %v", err)
	}
}

func writeArtifactFixture(t *testing.T, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "artifact.json")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write artifact fixture: %v", err)
	}
	return path
}
