package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeMaterialLoadsLocalProfiles(t *testing.T) {
	cases := []struct {
		profile string
		mode    string
	}{
		{profile: "single-tenant", mode: "onpremise"},
		{profile: "multi-tenant", mode: "saas"},
	}
	for _, tc := range cases {
		t.Run(tc.profile, func(t *testing.T) {
			root := t.TempDir()
			lines, err := runtimeMaterial(root, tc.profile)
			if err != nil {
				t.Fatal(err)
			}
			env := append(minimalRuntimeBase(t, root, tc.profile, tc.mode), lines...)
			assertUniqueEnvKeys(t, env)
			assertOwnerOnlyRuntimeFiles(t, root, tc.profile)
			if err = loadStartupProfile(t, "api", env); err != nil {
				t.Fatalf("api: %v", err)
			}
			if err = loadStartupProfile(t, "worker", env); err != nil {
				t.Fatalf("worker: %v", err)
			}
			loadOperatorStartup(t, tc.mode, lines, env)
		})
	}
}

func TestRuntimeMaterialRejectsUnknownProfile(t *testing.T) {
	if _, err := runtimeMaterial(t.TempDir(), "staging"); err == nil {
		t.Fatal("unknown profile was accepted")
	}
}

func loadOperatorStartup(t *testing.T, mode string, lines, env []string) {
	t.Helper()
	if mode == "saas" {
		if err := loadStartupProfile(t, "operator_initialize", env); err != nil {
			t.Fatalf("operator: %v", err)
		}
		return
	}
	requireEnvKey(t, lines, "HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH")
	requireEnvKey(t, lines, "HARPY_RELEASE_TRUST_ROOT_PATH")
	if err := loadStartupProfile(t, "operator_initialize", env); err != nil {
		t.Fatalf("operator: %v", err)
	}
	err := loadStartupProfilePreserving(t, "operator_initialize", env,
		"HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH",
		"HARPY_RELEASE_TRUST_ROOT_PATH",
	)
	if err == nil || !strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("on-prem operator accepted process authority keys: %v", err)
	}
}

func minimalRuntimeBase(t *testing.T, root, profile, mode string) []string {
	t.Helper()
	secrets := filepath.Join(root, "deploy", profile, "secrets")
	if err := os.MkdirAll(secrets, 0o700); err != nil {
		t.Fatal(err)
	}
	password := filepath.Join(secrets, "initial_admin_password")
	access := filepath.Join(secrets, "rustfs_access_key")
	secret := filepath.Join(secrets, "rustfs_secret_key")
	for path, body := range map[string]string{
		password: "local-admin-password\n",
		access:   "LOCALACCESSKEY1234567\n",
		secret:   "local-secret-key\n",
	} {
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return []string{
		"HARPY_ENVIRONMENT=development",
		"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_POSTGRES_DSN=postgresql://harpy:harpy@127.0.0.1:5432/harpy?sslmode=disable",
		"HARPY_LOG_LEVEL=info",
		"DEPLOYMENT_MODE=" + mode,
		"HARPY_PUBLIC_LISTEN_ADDR=127.0.0.1:8080",
		"HARPY_MANAGEMENT_LISTEN_ADDR=0.0.0.0:9090",
		"HARPY_RABBITMQ_URL=amqp://harpy:harpy@127.0.0.1:5672/",
		"GOMEMLIMIT=512MiB",
		"HARPY_INITIAL_ADMIN_PASSWORD_FILE=" + password,
		"HARPY_DOCUMENTS_S3_ENDPOINT=http://127.0.0.1:9000",
		"HARPY_DOCUMENTS_S3_REGION=us-east-1",
		"HARPY_DOCUMENTS_S3_BUCKET=harpy",
		"HARPY_DOCUMENTS_S3_PATH_STYLE=true",
		"HARPY_DOCUMENTS_S3_CREDENTIAL_SOURCE=static_files",
		"HARPY_DOCUMENTS_S3_ACCESS_KEY_ID_FILE=" + access,
		"HARPY_DOCUMENTS_S3_SECRET_ACCESS_KEY_FILE=" + secret,
	}
}
