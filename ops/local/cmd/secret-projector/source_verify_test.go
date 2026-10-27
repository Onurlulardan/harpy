package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestVerifySourceAcceptsExactHashedCatalog(t *testing.T) {
	now := time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC)
	root := writeTestSource(t, now)
	verified, err := verifySource(root, now)
	if err != nil {
		t.Fatal(err)
	}
	if verified.generation == "" || len(verified.files) == 0 {
		t.Fatal("verified source is incomplete")
	}
}

func TestVerifySourceRejectsTamperedArtifact(t *testing.T) {
	now := time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC)
	root := writeTestSource(t, now)
	path := filepath.Join(root, "postgres-runtime-dsn.txt")
	if err := os.WriteFile(path, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verifySource(root, now); err == nil {
		t.Fatal("tampered artifact was accepted")
	}
}

func TestVerifySourceRejectsUncatalogedEntry(t *testing.T) {
	now := time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC)
	root := writeTestSource(t, now)
	if err := os.WriteFile(filepath.Join(root, "extra"), []byte("extra"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verifySource(root, now); err == nil {
		t.Fatal("uncataloged entry was accepted")
	}
}

func TestVerifySourceRejectsRenewalDeadline(t *testing.T) {
	now := time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC)
	root := writeTestSource(t, now)
	if _, err := verifySource(root, now.Add(24*24*time.Hour)); err == nil {
		t.Fatal("source past renew-after was accepted")
	}
}

func TestVerifySourceRejectsDifferentArtifactContract(t *testing.T) {
	now := time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC)
	root := writeTestSource(t, now)
	path := filepath.Join(root, "manifest.json")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	updated := []byte(strings.Replace(
		string(content),
		fmt.Sprintf(`"artifact_contract_revision":%d`, supportedArtifactContractRevision),
		`"artifact_contract_revision":4`, 1,
	))
	if err := os.WriteFile(path, updated, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := verifySource(root, now); err == nil {
		t.Fatal("different artifact contract revision was accepted")
	}
}
