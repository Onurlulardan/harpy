package cutover

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestAuthorityArtifactBoundAcceptsMaximumAndRejectsNextByte(t *testing.T) {
	directory := t.TempDir()
	maximumPath := filepath.Join(directory, "authority-maximum.json")
	if err := os.WriteFile(
		maximumPath, bytes.Repeat([]byte("a"), maximumCutoverAuthorityBytes), 0o600,
	); err != nil {
		t.Fatal(err)
	}
	document, err := readArtifact(maximumPath, maximumCutoverAuthorityBytes)
	if err != nil || len(document) != maximumCutoverAuthorityBytes {
		t.Fatalf("maximum authority artifact rejected: %v", err)
	}
	overflowPath := filepath.Join(directory, "authority-overflow.json")
	if err := os.WriteFile(
		overflowPath, bytes.Repeat([]byte("b"), maximumCutoverAuthorityBytes+1), 0o600,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := readArtifact(overflowPath, maximumCutoverAuthorityBytes); err == nil {
		t.Fatal("authority artifact above exact cap accepted")
	}
}
