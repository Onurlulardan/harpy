package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComposeProfilesIncludePinnedStack(t *testing.T) {
	root := moduleRoot(t)
	third := mustRead(t, filepath.Join(root, "deploy", "third-party.yml"))
	single := mustRead(t, filepath.Join(root, "deploy", "single-tenant", "compose.yml"))
	multi := mustRead(t, filepath.Join(root, "deploy", "multi-tenant", "compose.yml"))
	if strings.ReplaceAll(multi, "TENANCY_MODE: multi", "TENANCY_MODE: single") != single {
		t.Fatal("profiles differ by more than TENANCY_MODE")
	}
	if strings.Contains(single, "image:") || strings.Contains(multi, "image:") {
		t.Fatal("profile redeclares an image")
	}
	if !strings.Contains(single, "../third-party.yml") || !strings.Contains(single, "TENANCY_MODE: single") || !strings.Contains(multi, "TENANCY_MODE: multi") {
		t.Fatal("profile include or tenancy mode missing")
	}
	for _, needle := range []string{
		"pgvector/pgvector:0.8.3-pg18-bookworm@sha256:c8a919765f2ef63681329fa21021b830cd4d79d1165bdca730dd016014e4da84",
		"rabbitmq:4.3.2-management-alpine@sha256:0753b75ce99094c385483d89449d532a0544fb85e4942a478b21cc497ab66d33",
		"rustfs/rustfs:1.0.4@sha256:ed9210d6411d77ffceb70d94b964a3cb7d225983368bc25bebe01e995293dcc2",
		`command: ["/data"]`,
		"RUSTFS_ACCESS_KEY_FILE: /run/secrets/rustfs_access_key",
		"RUSTFS_SECRET_KEY_FILE: /run/secrets/rustfs_secret_key",
		"http://127.0.0.1:9000/health",
		`"127.0.0.1:9000:9000"`,
		`"127.0.0.1:9001:9001"`,
		"rustfs-data:/data",
		"RABBITMQ_CONFIG_FILE: /run/secrets/rabbitmq",
		"target: /run/secrets/rabbitmq.conf",
		"POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password",
	} {
		if !strings.Contains(third, needle) {
			t.Fatalf("third-party.yml missing %s", needle)
		}
	}
	if strings.Contains(third, "RUSTFS_ACCESS_KEY:") || strings.Contains(third, "RUSTFS_SECRET_KEY:") {
		t.Fatal("rustfs direct credential env is set")
	}
	lower := strings.ToLower(third)
	for _, word := range []string{"minio", "grafana", "prometheus", "jaeger", "latest", "{0...3}", "rustfsadmin"} {
		if strings.Contains(lower, word) {
			t.Fatalf("third-party.yml contains %s", word)
		}
	}
	if strings.Count(third, "image:") != 3 || strings.Count(third, "\n  postgres:\n") != 1 || strings.Count(third, "\n  rabbitmq:\n") != 1 || strings.Count(third, "\n  rustfs:\n") != 1 {
		t.Fatal("third-party.yml does not define the three services once")
	}
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
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
