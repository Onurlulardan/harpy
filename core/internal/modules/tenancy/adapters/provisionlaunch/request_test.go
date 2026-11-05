package provisionlaunch

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestRunWritesCanonicalRequestAndReturnsPending(t *testing.T) {
	path := useRequestPath(t)
	err := Run(context.Background(), "00112233-4455-6677-8899-AABBCCDDEEFF")
	if !errors.Is(err, ErrPending) {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	const want = "{\"tenant_id\":\"00112233-4455-6677-8899-aabbccddeeff\"}\n"
	if string(body) != want {
		t.Fatalf("request = %q", body)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o", info.Mode().Perm())
	}
}

func TestRunRejectsInvalidCommandWithoutWriting(t *testing.T) {
	path := useRequestPath(t)
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	cases := []struct {
		ctx context.Context
		id  string
	}{
		{context.Background(), ""},
		{context.Background(), "not-a-uuid"},
		{context.Background(), uuid.Nil.String()},
		{nil, "00112233-4455-6677-8899-aabbccddeeff"},
		{cancelled, "00112233-4455-6677-8899-aabbccddeeff"},
	}
	for _, test := range cases {
		err := Run(test.ctx, test.id)
		if err == nil || errors.Is(err, ErrPending) ||
			err.Error() != "tenant provisioning command is invalid" {
			t.Fatalf("id=%q err=%v", test.id, err)
		}
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Fatalf("id=%q created %s (%v)", test.id, path, statErr)
		}
	}
}

func TestReadTenantIDRoundTrip(t *testing.T) {
	useRequestPath(t)
	const tenantID = "00112233-4455-6677-8899-aabbccddeeff"
	if err := Run(context.Background(), tenantID); !errors.Is(err, ErrPending) {
		t.Fatal(err)
	}
	got, err := ReadTenantID()
	if err != nil || got != tenantID {
		t.Fatalf("id=%q err=%v", got, err)
	}
}

func TestReadTenantIDRejectsMalformedRequest(t *testing.T) {
	path := useRequestPath(t)
	bodies := []string{
		"{\"tenant_id\":\"00112233-4455-6677-8899-aabbccddeeff\",\"extra\":true}\n",
		"{\"tenant_id\":\"\"}\n",
		"{\"tenant_id\":\"00112233-4455-6677-8899-AABBCCDDEEFF\"}\n",
		"{\"tenant_id\":\"00112233445566778899aabbccddeeff\"}\n",
		"{\"tenant_id\":\"00000000-0000-0000-0000-000000000000\"}\n",
	}
	for _, body := range bodies {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := ReadTenantID()
		if err == nil || errors.Is(err, ErrPending) || got != "" {
			t.Fatalf("body=%s id=%q err=%v", body, got, err)
		}
	}
}

func TestRequestSourceDoesNotExecuteDocker(t *testing.T) {
	source := "core/internal/modules/tenancy/adapters/provisionlaunch/request.go"
	body, err := os.ReadFile(filepath.Join(moduleRoot(t), source))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, forbidden := range []string{"exec.Command", "os/exec", "docker"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("request.go contains %s", forbidden)
		}
	}
}

func TestRuntimeComposeKeepsProvisioningOneShot(t *testing.T) {
	body, err := os.ReadFile(filepath.Join(moduleRoot(t), "ops/local/compose.runtime.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "restart: \"no\"") ||
		!strings.Contains(text, "env_file: ./state/secrets/tenant-provisioning.env") ||
		!strings.Contains(text, "entrypoint: [\"/app/core-initialize\"]") {
		t.Fatal("tenant-provisioning is not a one-shot initialize service")
	}
	// PRESERVE_NON_MIGRATION: absence guard for removed executor service names.
	for _, forbidden := range []string{
		"HARPY_GOOSE_PASSWORD", "HARPY_FOUNDATION_ADMIN",
		"tenant-migration-executor", "schema-executor", "core-migrator",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("compose.runtime.yml contains %s", forbidden)
		}
	}
	if strings.Count(text, "env_file: ./state/secrets/api.env") != 1 ||
		strings.Count(text, "env_file: ./state/secrets/worker.env") != 1 {
		t.Fatal(text)
	}
}

func useRequestPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "provision-request.json")
	previous := RequestPath
	RequestPath = path
	t.Cleanup(func() { RequestPath = previous })
	return path
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}
