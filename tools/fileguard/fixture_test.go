package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFixture(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func repeatLines(line string, count int) string {
	return strings.Repeat(line+"\n", count)
}

func containsViolation(violations []string, want string) bool {
	for _, violation := range violations {
		if strings.Contains(violation, want) {
			return true
		}
	}
	return false
}

func inspectFixture(t *testing.T, root string) []string {
	t.Helper()
	violations, err := inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	return violations
}

func writeManifest(t *testing.T, root, rel, content string) {
	t.Helper()
	sum := sha256.Sum256([]byte(content))
	body := fmt.Sprintf(
		"{\"version\":1,\"files\":[{\"path\":%q,\"generator\":\"test\",\"input_sha256\":%q,\"output_sha256\":%q}]}\n",
		rel, strings.Repeat("ab", 32), hex.EncodeToString(sum[:]))
	writeFixture(t, root, "contracts/generated.yaml", body)
}
