package runtimev1

func addFormsExportProfileSnapshot(variables map[string]any, worker bool) {
	variables["HARPY_FORMS_EXPORT_S3_ENDPOINT"] = "https://s3.fixture.invalid"
	variables["HARPY_FORMS_EXPORT_S3_REGION"] = "eu-central-1"
	variables["HARPY_FORMS_EXPORT_S3_BUCKET"] = "harpy-form-exports"
	variables["HARPY_FORMS_EXPORT_S3_EXPECTED_BUCKET_OWNER"] = "000000000000"
	variables["HARPY_FORMS_EXPORT_S3_PATH_STYLE"] = "true"
	variables["HARPY_FORMS_EXPORT_S3_CREDENTIAL_SOURCE"] = "workload_identity"
	variables["HARPY_FORMS_EXPORT_FCART_KEK_KEYRING_FILE"] = "/secrets/forms-export-fcart-keyring.json"
	if worker {
		variables["HARPY_FORMS_EXPORT_SPOOL_DIR"] = "/var/lib/harpy/forms-export-spool"
	}
}
