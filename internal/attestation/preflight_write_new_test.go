package attestation

import (
	"path/filepath"
	"testing"
)

func TestPreflightWriteNewDirectoryRequiresAbsoluteExistingDirectory(t *testing.T) {
	directory := t.TempDir()
	if err := PreflightWriteNewDirectory(directory); err != nil {
		t.Fatal(err)
	}
	if err := PreflightWriteNewDirectory("relative"); err == nil {
		t.Fatal("relative publication directory was accepted")
	}
	if err := PreflightWriteNewDirectory(filepath.Join(directory, "missing")); err == nil {
		t.Fatal("missing publication directory was accepted")
	}
}
