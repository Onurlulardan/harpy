package configfixture

import (
	"fmt"
	"os"
	"path/filepath"
)

func setIAMSystemTransfer(setter interface {
	EnvironmentSetter
	ArtifactFixture
}) {
	for _, item := range iamSystemTransferEntries(setter) {
		setter.Setenv(item.name, item.value)
	}
}

func iamSystemTransferEnvironment(fixture ArtifactFixture) []string {
	return environmentStrings(iamSystemTransferEntries(fixture))
}

func iamSystemTransferEntries(fixture ArtifactFixture) []entry {
	fixture.Helper()
	directory := fixture.TempDir()
	spool := filepath.Join(directory, "spool")
	if err := os.Mkdir(spool, 0o700); err != nil {
		fixture.Fatalf("create IAM system transfer spool fixture: %v", err)
	}
	exportKeys := writeIAMTransferKeyring(fixture, directory,
		"export-keyring.json", "iam-system-export-fixture", 111)
	importKeys := writeIAMTransferKeyring(fixture, directory,
		"import-keyring.json", "iam-system-import-fixture", 121)
	return []entry{
		{"HARPY_IAM_SYSTEM_TRANSFER_SPOOL_DIR", spool},
		{"HARPY_IAM_SYSTEM_EXPORT_S3_ENDPOINT", "https://s3.iam-export.fixture.invalid"},
		{"HARPY_IAM_SYSTEM_EXPORT_S3_REGION", "eu-central-1"},
		{"HARPY_IAM_SYSTEM_EXPORT_S3_BUCKET", "harpy-iam-system-exports"},
		{"HARPY_IAM_SYSTEM_EXPORT_S3_EXPECTED_BUCKET_OWNER", "000000000000"},
		{"HARPY_IAM_SYSTEM_EXPORT_S3_PATH_STYLE", "true"},
		{"HARPY_IAM_SYSTEM_EXPORT_S3_CREDENTIAL_SOURCE", "workload_identity"},
		{"HARPY_IAM_SYSTEM_EXPORT_FCART_KEK_KEYRING_FILE", exportKeys},
		{"HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_ENDPOINT", "https://s3.iam-import.fixture.invalid"},
		{"HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_REGION", "eu-central-1"},
		{"HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_BUCKET", "harpy-iam-system-import-sources"},
		{"HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_EXPECTED_BUCKET_OWNER", "000000000000"},
		{"HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_PATH_STYLE", "true"},
		{"HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_CREDENTIAL_SOURCE", "workload_identity"},
		{"HARPY_IAM_SYSTEM_IMPORT_SOURCE_FCART_KEK_KEYRING_FILE", importKeys},
	}
}

func writeIAMTransferKeyring(fixture ArtifactFixture, directory, name, id string,
	marker byte) string {
	path := filepath.Join(directory, name)
	document := fmt.Sprintf(
		`{"version":1,"current":{"kid":%q,"key":%q},"retained":[]}`,
		id, material(marker))
	if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
		fixture.Fatalf("create IAM system transfer keyring fixture: %v", err)
	}
	return path
}
