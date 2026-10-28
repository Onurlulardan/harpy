package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

const expectedEnvLocal = `COMPOSE_PROJECT_NAME=harpy-local
LOCAL_API_PORT=8000
LOCAL_API_MANAGEMENT_PORT=9001
LOCAL_WORKER_MANAGEMENT_PORT=9002
LOCAL_SCHEMA_MANAGEMENT_PORT=9003
LOCAL_TENANT_MIGRATION_MANAGEMENT_PORT=9004
LOCAL_POSTGRES_PORT=55432
LOCAL_RABBITMQ_PORT=5673
LOCAL_RABBITMQ_MANAGEMENT_PORT=15673
LOCAL_S3_PORT=19000
LOCAL_MAILPIT_SMTP_PORT=1026
LOCAL_MAILPIT_UI_PORT=8026
LOCAL_PUBLIC_BASE_URL=http://localhost:5173
`

func TestEnvLocalWritesMissingFileAndLeavesCustomPort(t *testing.T) {
	repoBefore := readWorkingModuleEnvLocal(t)

	root, start := tempModuleStart(t)
	if err := ensureEnvLocalFrom(start); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(filepath.Join(root, envLocalFileName))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(written, []byte(expectedEnvLocal)) {
		t.Fatalf("missing .env.local was not the Harpy compose file:\n%s", written)
	}
	if _, err := os.Lstat(filepath.Join(start, envLocalFileName)); !os.IsNotExist(err) {
		t.Fatalf(".env.local was written below the module root: %v", err)
	}

	existing := []byte("LOCAL_API_PORT=8123\n")
	kept := filepath.Join(root, envLocalFileName)
	if err := os.WriteFile(kept, existing, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ensureEnvLocalFrom(start); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(kept)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, existing) {
		t.Fatalf("existing LOCAL_API_PORT=8123 was changed:\n%s", after)
	}

	repoAfter := readWorkingModuleEnvLocal(t)
	if !bytes.Equal(repoBefore, repoAfter) {
		t.Fatal("repository .env.local was modified")
	}
}

func tempModuleStart(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(root, "go.mod"),
		[]byte("module example.com/envlocal\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	start := filepath.Join(root, "cmd", "work")
	if err := os.MkdirAll(start, 0o700); err != nil {
		t.Fatal(err)
	}
	return root, start
}

func readWorkingModuleEnvLocal(t *testing.T) []byte {
	t.Helper()
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := moduleRoot(workingDirectory)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(root, envLocalFileName))
	if err != nil {
		t.Fatal(err)
	}
	return body
}
