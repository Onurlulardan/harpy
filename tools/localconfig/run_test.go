package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunCLIResetRemovesOnlyVerifiedArtifactSet(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	output := filepath.Join(t.TempDir(), "secrets")
	if _, err := ensureArtifactSet(output, now); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runCLI([]string{"-output", output, "-reset"},
		&stdout, &stderr, func() time.Time { return now.Add(24 * 24 * time.Hour) })
	if exitCode != 0 || stderr.Len() != 0 || !strings.Contains(stdout.String(), "reset") {
		t.Fatalf("reset CLI failed: exit=%d stdout=%q stderr=%q",
			exitCode, stdout.String(), stderr.String())
	}
	if _, err := os.Lstat(output); !os.IsNotExist(err) {
		t.Fatalf("reset output still exists: %v", err)
	}
}

func TestRunCLIRequiresOutput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := runCLI(nil, &stdout, &stderr, time.Now)
	if exitCode != 2 || !strings.Contains(stderr.String(), "usage:") {
		t.Fatalf("missing output accepted: exit=%d stderr=%q", exitCode, stderr.String())
	}
}
