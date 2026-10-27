package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureModulePath = "github.com/Onurlulardan/harpy"

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

func goImports(paths ...string) string {
	var builder strings.Builder
	builder.WriteString("package sample\n\nimport (\n")
	for _, path := range paths {
		fmt.Fprintf(&builder, "\t_ %q\n", path)
	}
	builder.WriteString(")\n")
	return builder.String()
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
	goMod := filepath.Join(root, "go.mod")
	if _, err := os.Stat(goMod); err != nil {
		if writeErr := os.WriteFile(goMod, []byte("module "+fixtureModulePath+"\n"), 0o600); writeErr != nil {
			t.Fatal(writeErr)
		}
	}
	violations, err := inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	return violations
}
