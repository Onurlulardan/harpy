package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormsExportArtifactsLoadForAPIAndWorker(t *testing.T) {
	api, err := loadAPI(validAPIEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	artifacts, err := api.FormsExportArtifacts()
	if err != nil || artifacts.Validate() != nil {
		t.Fatalf("API Forms export artifacts invalid: %v", err)
	}
	storage, _ := artifacts.Storage()
	endpoint, region, bucket, owner, pathStyle, err := storage.Connection()
	if err != nil || endpoint != "https://s3.fixture.invalid" || region != "eu-central-1" ||
		bucket != "harpy-form-exports" || owner != "000000000000" || !pathStyle {
		t.Fatalf("unexpected S3 connection: %q %q %q %q %t %v", endpoint, region, bucket, owner, pathStyle, err)
	}
	credentials, _ := storage.Credentials()
	source, err := credentials.Source()
	if err != nil || source != FormsExportWorkloadIdentity {
		t.Fatalf("unexpected credential source: %q %v", source, err)
	}

	worker, err := loadWorker(validWorkerEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := worker.FormsExportArtifacts(); err != nil ||
		worker.FormsExportSpoolDirectory().Value() == "" {
		t.Fatalf("worker Forms export config invalid: %v", err)
	}
}

func TestFormsExportStaticCredentialsLoadOnlyFromFiles(t *testing.T) {
	directory := t.TempDir()
	accessPath := writeFormsExportSecret(t, directory, "access", "fixture-access-id\n")
	secretPath := writeFormsExportSecret(t, directory, "secret", "fixture-secret-key\r\n")
	tokenPath := writeFormsExportSecret(t, directory, "token", "fixture-session-token")
	entries := withEnvironment(
		validAPIEnvironment(),
		formsExportCredentialSourceKey+"=static_files",
		formsExportAccessKeyFileKey+"="+accessPath,
		formsExportSecretKeyFileKey+"="+secretPath,
		formsExportSessionTokenFileKey+"="+tokenPath,
	)
	loaded, err := loadAPI(entries)
	if err != nil {
		t.Fatal(err)
	}
	artifacts, _ := loaded.FormsExportArtifacts()
	storage, _ := artifacts.Storage()
	credentials, _ := storage.Credentials()
	access, secret, token, hasToken, err := credentials.Static()
	if err != nil || !hasToken || access.Value() != "fixture-access-id" ||
		secret.Value() != "fixture-secret-key" || token.Value() != "fixture-session-token" {
		t.Fatalf("unexpected static credentials: %v", err)
	}
}

func TestFormsExportKeysAreProfileScoped(t *testing.T) {
	if _, err := loadAPI(append(
		validAPIEnvironment(), "HARPY_FORMS_EXPORT_SPOOL_DIR="+t.TempDir(),
	)); err == nil {
		t.Fatal("API accepted worker-only spool directory")
	}
	if _, err := loadAPI(append(
		validAPIEnvironment(), "HARPY_FORMS_EXPORT_S3_ACCESS_KEY_ID=secret-value",
	)); err == nil {
		t.Fatal("API accepted a direct static credential value")
	}
}

func writeFormsExportSecret(t *testing.T, directory, name, value string) string {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
