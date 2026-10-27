package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalEnvMatchesProcessKeys(t *testing.T) {
	root := t.TempDir()
	first := runOK(t, root, []string{"--defaults"}, "")
	env := readEnv(t, root, "single-tenant")
	if env["DEPLOYMENT_MODE"] != "onpremise" || env["HARPY_ENVIRONMENT"] != "development" ||
		env["HARPY_PUBLIC_LISTEN_ADDR"] != "127.0.0.1:8080" ||
		env["HARPY_MANAGEMENT_LISTEN_ADDR"] != localManagementListenAddress ||
		env["HARPY_POSTGRES_DSN"] == "" || env["HARPY_RABBITMQ_URL"] == "" ||
		env["GOMEMLIMIT"] != "512MiB" || env["HARPY_POSTGRES_ENDPOINT_KIND"] != "direct_postgres" {
		t.Fatalf("process env: %#v", env)
	}
	if _, found := env["HARPY_DOCUMENTS_CLAMAV_ADDRESS"]; found {
		t.Fatal("local profile requires ClamAV")
	}
	if _, found := env["HARPY_DOCUMENTS_S3_EXPECTED_BUCKET_OWNER"]; found {
		t.Fatal("local profile requires a bucket owner")
	}
	body := mustRead(t, filepath.Join(root, "deploy", "single-tenant", ".env"))
	if strings.Contains(body, "gcp-kms://") {
		t.Fatal("local profile requires GCP KMS")
	}
	adminPath := filepath.Join(root, "deploy", "single-tenant", "secrets", "initial_admin_password")
	admin := strings.TrimRight(mustRead(t, adminPath), "\r\n")
	if admin == "" || strings.Contains(first, admin) || !strings.Contains(body, "HARPY_INITIAL_ADMIN_PASSWORD_FILE=") {
		t.Fatal("admin password missing or printed")
	}
	if !strings.Contains(body, "HARPY_DOCUMENTS_S3_PATH_STYLE=true") ||
		!strings.Contains(body, "HARPY_DOCUMENTS_S3_CREDENTIAL_SOURCE=static_files") {
		t.Fatal("rustfs documents credentials were not written")
	}
	_ = runOK(t, root, []string{"--defaults"}, "")
	if strings.TrimRight(mustRead(t, adminPath), "\r\n") != admin {
		t.Fatal("second run rotated the admin password")
	}
	aws := "multi-tenant\n9090\n5433\nappdb\nappuser\nexternal\nhttps://s3.eu-west-1.amazonaws.com\naws\neu-west-1\nmybucket\nAKIAEXAMPLEKEY\nawssecretvalue\n"
	runOK(t, root, nil, aws)
	awsEnv := readEnv(t, root, "multi-tenant")
	if awsEnv["DEPLOYMENT_MODE"] != "saas" || awsEnv["HARPY_PUBLIC_LISTEN_ADDR"] != "127.0.0.1:9090" {
		t.Fatalf("aws process env: %#v", awsEnv)
	}
	if _, found := awsEnv["HARPY_DOCUMENTS_S3_ENDPOINT"]; found {
		t.Fatal("aws profile wrote documents keys that require a bucket owner")
	}
}

func TestSingleTenantEnvWritesBootMaterial(t *testing.T) {
	root := t.TempDir()
	runOK(t, root, []string{"--defaults"}, "")
	env := readEnv(t, root, "single-tenant")
	cursor := env["HARPY_REST_QUERY_CURSOR_KEYRING"]
	if cursor == "" {
		cursor = env["HARPY_REST_QUERY_CURSOR_KEYRING_FILE"]
	}
	manifest := env["HARPY_RELEASE_MANIFEST_PATH"]
	if cursor == "" || manifest == "" {
		t.Fatalf("cursor=%q manifest=%q", cursor, manifest)
	}
	if _, err := os.Stat(cursor); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(manifest); err != nil {
		t.Fatal(err)
	}
	if env["HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH"] == "" || env["HARPY_RELEASE_TRUST_ROOT_PATH"] == "" {
		t.Fatal("shared authority paths missing")
	}
	assertEnvFileKeysUnique(t, root, "single-tenant")
	_ = runOK(t, root, []string{"--defaults"}, "")
	again := readEnv(t, root, "single-tenant")
	againCursor := again["HARPY_REST_QUERY_CURSOR_KEYRING"]
	if againCursor == "" {
		againCursor = again["HARPY_REST_QUERY_CURSOR_KEYRING_FILE"]
	}
	if againCursor == "" || again["HARPY_RELEASE_MANIFEST_PATH"] == "" {
		t.Fatal("second run dropped boot material")
	}
	if _, err := os.Stat(againCursor); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(again["HARPY_RELEASE_MANIFEST_PATH"]); err != nil {
		t.Fatal(err)
	}
}

func assertEnvFileKeysUnique(t *testing.T, root, profile string) {
	t.Helper()
	seen := map[string]struct{}{}
	body := mustRead(t, filepath.Join(root, "deploy", profile, ".env"))
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(strings.TrimRight(line, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, _, ok := strings.Cut(line, "=")
		if !ok || key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate %s", key)
		}
		seen[key] = struct{}{}
	}
}
