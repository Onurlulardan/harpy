package runtimev1

func addAnalyticsDashboardExportProfileSnapshot(
	variables map[string]any, worker bool,
) {
	variables["HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_ENDPOINT"] = "https://s3.fixture.invalid"
	variables["HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_REGION"] = "eu-central-1"
	variables["HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_BUCKET"] = "harpy-analytics-dashboard-exports"
	variables["HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_EXPECTED_BUCKET_OWNER"] = "000000000000"
	variables["HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_PATH_STYLE"] = "true"
	variables["HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_CREDENTIAL_SOURCE"] = "workload_identity"
	variables["HARPY_ANALYTICS_DASHBOARD_EXPORT_FCART_KEK_KEYRING_FILE"] = "/secrets/analytics-export-fcart-keyring.json"
	if worker {
		variables["HARPY_ANALYTICS_DASHBOARD_EXPORT_SPOOL_DIR"] =
			"/var/lib/harpy/analytics-export-spool"
	}
}
