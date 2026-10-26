package configfixture

import (
	"fmt"
	"os"
	"path/filepath"
)

func setFormsImportSource(setter interface {
	EnvironmentSetter
	ArtifactFixture
}, api bool) {
	for _, item := range formsImportSourceEntries(setter, api) {
		setter.Setenv(item.name, item.value)
	}
}

func formsImportSourceEnvironment(fixture ArtifactFixture, api bool) []string {
	return environmentStrings(formsImportSourceEntries(fixture, api))
}

func formsImportSourceEntries(fixture ArtifactFixture, api bool) []entry {
	fixture.Helper()
	directory := fixture.TempDir()
	keyringPath := filepath.Join(directory, "forms-import-source-fcart-keyring.json")
	keyring := fmt.Sprintf(
		`{"version":1,"current":{"kid":"forms-import-source-fixture","key":"%s"},"retained":[]}`,
		material(91))
	if err := os.WriteFile(keyringPath, []byte(keyring), 0o600); err != nil {
		fixture.Fatalf("create Forms import source keyring fixture: %v", err)
	}
	entries := []entry{
		{"HARPY_FORMS_IMPORT_SOURCE_S3_ENDPOINT", "https://s3.fixture.invalid"},
		{"HARPY_FORMS_IMPORT_SOURCE_S3_REGION", "eu-central-1"},
		{"HARPY_FORMS_IMPORT_SOURCE_S3_BUCKET", "harpy-form-import-sources"},
		{"HARPY_FORMS_IMPORT_SOURCE_S3_EXPECTED_BUCKET_OWNER", "000000000000"},
		{"HARPY_FORMS_IMPORT_SOURCE_S3_PATH_STYLE", "true"},
		{"HARPY_FORMS_IMPORT_SOURCE_S3_CREDENTIAL_SOURCE", "workload_identity"},
		{"HARPY_FORMS_IMPORT_SOURCE_FCART_KEK_KEYRING_FILE", keyringPath},
	}
	if api {
		staging := filepath.Join(directory, "staging")
		if err := os.Mkdir(staging, 0o700); err != nil {
			fixture.Fatalf("create Forms import source staging fixture: %v", err)
		}
		entries = append(entries,
			entry{"HARPY_FORMS_IMPORT_SOURCE_STAGING_DIR", staging})
	}
	return entries
}
