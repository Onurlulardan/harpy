package config

import (
	"strings"
	"testing"
)

func TestFormsImportSourceLoadsForAPIAndWorker(t *testing.T) {
	api, err := loadAPI(validAPIEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	settings, err := api.FormsImportSourceArtifacts()
	if err != nil || settings.Validate() != nil ||
		api.FormsImportSourceStagingDirectory().Value() == "" {
		t.Fatalf("API import-source config = %#v, %v", settings, err)
	}
	worker, err := loadWorker(validWorkerEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	settings, err = worker.FormsImportSourceArtifacts()
	if err != nil || settings.Validate() != nil {
		t.Fatalf("worker import-source config = %#v, %v", settings, err)
	}
}

func TestFormsImportSourceRequiresTLSInProduction(t *testing.T) {
	entries := withEnvironment(validAPIEnvironment(),
		"HARPY_ENVIRONMENT=production",
		"HARPY_POSTGRES_DSN=postgresql://user@example.invalid/harpy?sslmode=verify-full",
		formsImportSourceEndpointKey+"=http://s3.fixture.invalid",
		"HARPY_OTLP_ENDPOINT=https://otel.fixture.invalid",
		"HARPY_OTLP_INSECURE=false")
	if _, err := loadAPI(entries); err == nil ||
		!strings.Contains(err.Error(), formsImportSourceEndpointKey) {
		t.Fatalf("plaintext production endpoint error = %v", err)
	}
}

func TestFormsImportSourceRejectsDirectSecretsAndMissingStaging(t *testing.T) {
	direct := append(validAPIEnvironment(),
		"HARPY_FORMS_IMPORT_SOURCE_S3_SECRET_ACCESS_KEY=must-not-leak")
	if _, err := loadAPI(direct); err == nil || strings.Contains(err.Error(), "must-not-leak") {
		t.Fatalf("direct import-source secret error = %v", err)
	}
	missing := withoutEnvironmentKey(validAPIEnvironment(), formsImportSourceStagingKey)
	if _, err := loadAPI(missing); err == nil ||
		!strings.Contains(err.Error(), formsImportSourceStagingKey) {
		t.Fatalf("missing import-source staging error = %v", err)
	}
}
