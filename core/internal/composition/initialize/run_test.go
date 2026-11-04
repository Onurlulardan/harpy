package initialize

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunRejectsExtraArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"deploy", "up"}, &stdout, &stderr, func(string) string {
		return "postgres-super-secret-value"
	}, func(string) ([]byte, error) {
		return []byte("postgres-super-secret-value"), nil
	})
	if code != 2 || strings.Contains(stdout.String()+stderr.String(), "postgres-super-secret-value") {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "usage: core-initialize") {
		t.Fatal(stderr.String())
	}
}

func TestRunMissingConfigurationDoesNotOpenDatabase(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run(nil, &stdout, &stderr, func(string) string { return "" }, func(string) ([]byte, error) {
		t.Fatal("read a secret file")
		return nil, nil
	})
	if code != 2 || !strings.Contains(stderr.String(), "foundation admin dsn") {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
}
