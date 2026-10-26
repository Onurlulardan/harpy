package runtimev1

func addFormsImportSourceProfileSnapshot(variables map[string]any, api bool) {
	variables["HARPY_FORMS_IMPORT_SOURCE_S3_ENDPOINT"] = "https://s3.example.invalid"
	variables["HARPY_FORMS_IMPORT_SOURCE_S3_REGION"] = "eu-central-1"
	variables["HARPY_FORMS_IMPORT_SOURCE_S3_BUCKET"] = "harpy-form-import-sources"
	variables["HARPY_FORMS_IMPORT_SOURCE_S3_EXPECTED_BUCKET_OWNER"] = "000000000000"
	variables["HARPY_FORMS_IMPORT_SOURCE_S3_PATH_STYLE"] = "true"
	variables["HARPY_FORMS_IMPORT_SOURCE_S3_CREDENTIAL_SOURCE"] = "workload_identity"
	variables["HARPY_FORMS_IMPORT_SOURCE_FCART_KEK_KEYRING_FILE"] =
		"/secrets/forms-import-source-fcart-keyring.json"
	if api {
		variables["HARPY_FORMS_IMPORT_SOURCE_STAGING_DIR"] = "/var/lib/harpy/import-source"
	}
}
