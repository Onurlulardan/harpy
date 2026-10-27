package main

import (
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
