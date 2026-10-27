package configfixture

import (
	"fmt"
	"os"
	"path/filepath"
)

func setAnalyticsDashboardExport(setter interface {
	EnvironmentSetter
	ArtifactFixture
}, worker bool) {
	for _, item := range analyticsDashboardExportEntries(setter, worker) {
		setter.Setenv(item.name, item.value)
	}
}

func analyticsDashboardExportEnvironment(
	fixture ArtifactFixture, worker bool,
) []string {
	return environmentStrings(analyticsDashboardExportEntries(fixture, worker))
}

func analyticsDashboardExportEntries(
	fixture ArtifactFixture, worker bool,
) []entry {
	fixture.Helper()
	directory := fixture.TempDir()
	keyringPath := filepath.Join(directory, "analytics-dashboard-export-keyring.json")
	document := fmt.Sprintf(
		`{"version":1,"current":{"kid":"analytics-export-fixture","key":"%s"},"retained":[]}`,
		material(141),
	)
	if err := os.WriteFile(keyringPath, []byte(document), 0o600); err != nil {
		fixture.Fatalf("create Analytics export keyring fixture: %v", err)
	}
	entries := []entry{
		{"HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_ENDPOINT", "https://s3.fixture.invalid"},
		{"HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_REGION", "eu-central-1"},
		{"HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_BUCKET", "harpy-analytics-dashboard-exports"},
		{"HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_EXPECTED_BUCKET_OWNER", "000000000000"},
		{"HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_PATH_STYLE", "true"},
		{"HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_CREDENTIAL_SOURCE", "workload_identity"},
		{"HARPY_ANALYTICS_DASHBOARD_EXPORT_FCART_KEK_KEYRING_FILE", keyringPath},
	}
	if worker {
		spool := filepath.Join(directory, "spool")
		if err := os.Mkdir(spool, 0o700); err != nil {
			fixture.Fatalf("create Analytics export spool fixture: %v", err)
		}
		entries = append(entries,
			entry{"HARPY_ANALYTICS_DASHBOARD_EXPORT_SPOOL_DIR", spool})
	}
	return entries
}
