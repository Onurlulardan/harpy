package configfixture

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type APISetter interface {
	IAMAPISetter
}

func SetAPI(setter APISetter) {
	setter.Helper()
	setCatalogPackageKeyring(setter)
	SetIAMAPI(setter)
	setPhase14API(setter)
	setFormsExport(setter, false)
	setFormsImportSource(setter, true)
	setAnalyticsDashboardExport(setter, false)
	setIAMSystemTransfer(setter)
	setDocuments(setter, false)
	setNotifications(setter, false)
}

type WorkerSetter interface {
	IAMWorkerSetter
}

func SetWorker(setter WorkerSetter) {
	setter.Helper()
	setter.Setenv("GOMEMLIMIT", "512MiB")
	setCatalogPackageKeyring(setter)
	SetIAMWorker(setter)
	setPhase14Worker(setter)
	setRestQueryCursorKeyring(setter)
	setFormsExport(setter, true)
	setFormsImportSource(setter, false)
	setAnalyticsDashboardExport(setter, true)
	setIAMSystemTransfer(setter)
	setDocuments(setter, true)
	setNotifications(setter, true)
}

func APIEnvironment(fixture ArtifactFixture, now time.Time) []string {
	fixture.Helper()
	values := append([]string{catalogPackageKeyringEnvironment()},
		IAMAPIEnvironment(fixture, now)...)
	values = append(values, phase14APIEnvironment(fixture)...)
	values = append(values, formsExportEnvironment(fixture, false)...)
	values = append(values, formsImportSourceEnvironment(fixture, true)...)
	values = append(values, analyticsDashboardExportEnvironment(fixture, false)...)
	values = append(values, iamSystemTransferEnvironment(fixture)...)
	values = append(values, documentsEnvironment(fixture, false)...)
	values = append(values, notificationsEnvironment(fixture, false))
	return values
}

func WorkerEnvironment(fixture ArtifactFixture, now time.Time) []string {
	fixture.Helper()
	values := append([]string{
		"GOMEMLIMIT=512MiB", catalogPackageKeyringEnvironment(),
	}, IAMWorkerEnvironment(fixture, now)...)
	values = append(values, phase14WorkerEnvironment(fixture)...)
	values = append(values,
		restQueryCursorKeyringEnvironment())
	values = append(values, formsExportEnvironment(fixture, true)...)
	values = append(values, formsImportSourceEnvironment(fixture, false)...)
	values = append(values, analyticsDashboardExportEnvironment(fixture, true)...)
	values = append(values, iamSystemTransferEnvironment(fixture)...)
	values = append(values, documentsEnvironment(fixture, true)...)
	values = append(values, notificationsEnvironment(fixture, true))
	return values
}

func setFormsExport(setter interface {
	EnvironmentSetter
	ArtifactFixture
}, worker bool) {
	for _, item := range formsExportEntries(setter, worker) {
		setter.Setenv(item.name, item.value)
	}
}

func formsExportEnvironment(fixture ArtifactFixture, worker bool) []string {
	return environmentStrings(formsExportEntries(fixture, worker))
}

func formsExportEntries(fixture ArtifactFixture, worker bool) []entry {
	fixture.Helper()
	directory := fixture.TempDir()
	keyringPath := filepath.Join(directory, "forms-export-fcart-keyring.json")
	keyring := fmt.Sprintf(
		`{"version":1,"current":{"kid":"forms-export-fixture","key":"%s"},"retained":[]}`,
		material(71),
	)
	if err := os.WriteFile(keyringPath, []byte(keyring), 0o600); err != nil {
		fixture.Fatalf("create Forms export keyring fixture: %v", err)
	}
	entries := []entry{
		{"HARPY_FORMS_EXPORT_S3_ENDPOINT", "https://s3.fixture.invalid"},
		{"HARPY_FORMS_EXPORT_S3_REGION", "eu-central-1"},
		{"HARPY_FORMS_EXPORT_S3_BUCKET", "harpy-form-exports"},
		{"HARPY_FORMS_EXPORT_S3_EXPECTED_BUCKET_OWNER", "000000000000"},
		{"HARPY_FORMS_EXPORT_S3_PATH_STYLE", "true"},
		{"HARPY_FORMS_EXPORT_S3_CREDENTIAL_SOURCE", "workload_identity"},
		{"HARPY_FORMS_EXPORT_FCART_KEK_KEYRING_FILE", keyringPath},
	}
	if worker {
		spool := filepath.Join(directory, "spool")
		if err := os.Mkdir(spool, 0o700); err != nil {
			fixture.Fatalf("create Forms export spool fixture: %v", err)
		}
		entries = append(entries, entry{"HARPY_FORMS_EXPORT_SPOOL_DIR", spool})
	}
	return entries
}
