package configfixture

import (
	"fmt"
	"os"
	"path/filepath"
)

func setDocuments(setter interface {
	EnvironmentSetter
	ArtifactFixture
}, worker bool) {
	for _, item := range documentsEntries(setter, worker) {
		setter.Setenv(item.name, item.value)
	}
}

func documentsEnvironment(fixture ArtifactFixture, worker bool) []string {
	return environmentStrings(documentsEntries(fixture, worker))
}

func documentsEntries(fixture ArtifactFixture, worker bool) []entry {
	fixture.Helper()
	keyringPath := filepath.Join(fixture.TempDir(), "documents-kek-keyring.json")
	keyring := fmt.Sprintf(
		`{"version":1,"current":{"kid":"documents-fixture","key":"%s"},"retained":[]}`,
		material(101))
	if err := os.WriteFile(keyringPath, []byte(keyring), 0o600); err != nil {
		fixture.Fatalf("create Documents keyring fixture: %v", err)
	}
	entries := []entry{
		{"HARPY_DOCUMENTS_S3_ENDPOINT", "https://s3.documents.fixture.invalid"},
		{"HARPY_DOCUMENTS_S3_REGION", "eu-central-1"},
		{"HARPY_DOCUMENTS_S3_BUCKET", "harpy-documents"},
		{"HARPY_DOCUMENTS_S3_EXPECTED_BUCKET_OWNER", "000000000000"},
		{"HARPY_DOCUMENTS_S3_PATH_STYLE", "true"},
		{"HARPY_DOCUMENTS_S3_CREDENTIAL_SOURCE", "workload_identity"},
		{"HARPY_DOCUMENTS_KEK_KEYRING_FILE", keyringPath},
		{"HARPY_DOCUMENTS_BILLING_ACTOR_ID", "019b1234-5678-7abc-8def-0123456789ab"},
	}
	if !worker {
		return append(entries,
			entry{"HARPY_DOCUMENTS_CLAMAV_ADDRESS", "clamav.fixture.invalid:3310"})
	}
	return append(entries,
		entry{"HARPY_DOCUMENTS_CLAMAV_ADDRESS", "clamav.fixture.invalid:3310"},
		entry{"HARPY_DOCUMENTS_LEGACY_SOURCE_S3_ENDPOINT", "https://legacy.documents.fixture.invalid"},
		entry{"HARPY_DOCUMENTS_LEGACY_SOURCE_S3_REGION", "eu-central-1"},
		entry{"HARPY_DOCUMENTS_LEGACY_SOURCE_S3_BUCKET", "harpy-documents-legacy"},
		entry{"HARPY_DOCUMENTS_LEGACY_SOURCE_S3_EXPECTED_BUCKET_OWNER", "000000000000"},
		entry{"HARPY_DOCUMENTS_LEGACY_SOURCE_S3_PATH_STYLE", "true"},
		entry{"HARPY_DOCUMENTS_LEGACY_SOURCE_S3_CREDENTIAL_SOURCE", "workload_identity"},
		entry{"HARPY_DOCUMENTS_ARCHIVE_RESERVATION_TTL", "24h"},
	)
}
