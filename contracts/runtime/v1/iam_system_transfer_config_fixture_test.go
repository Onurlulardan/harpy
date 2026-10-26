package runtimev1

func addIAMSystemTransferProfileSnapshot(variables map[string]any) {
	variables["HARPY_IAM_SYSTEM_TRANSFER_SPOOL_DIR"] =
		"/var/lib/harpy/iam-system-transfer-spool"
	addIAMSystemExportProfileSnapshot(variables)
	addIAMSystemImportProfileSnapshot(variables)
}

func addIAMSystemExportProfileSnapshot(variables map[string]any) {
	variables["HARPY_IAM_SYSTEM_EXPORT_S3_ENDPOINT"] = "https://s3.iam-export.invalid"
	variables["HARPY_IAM_SYSTEM_EXPORT_S3_REGION"] = "eu-central-1"
	variables["HARPY_IAM_SYSTEM_EXPORT_S3_BUCKET"] = "harpy-iam-system-exports"
	variables["HARPY_IAM_SYSTEM_EXPORT_S3_EXPECTED_BUCKET_OWNER"] = "000000000000"
	variables["HARPY_IAM_SYSTEM_EXPORT_S3_PATH_STYLE"] = "true"
	variables["HARPY_IAM_SYSTEM_EXPORT_S3_CREDENTIAL_SOURCE"] = "workload_identity"
	variables["HARPY_IAM_SYSTEM_EXPORT_FCART_KEK_KEYRING_FILE"] =
		"/secrets/iam-system-export-fcart-keyring.json"
}

func addIAMSystemImportProfileSnapshot(variables map[string]any) {
	variables["HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_ENDPOINT"] = "https://s3.iam-import.invalid"
	variables["HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_REGION"] = "eu-central-1"
	variables["HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_BUCKET"] = "harpy-iam-system-imports"
	variables["HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_EXPECTED_BUCKET_OWNER"] = "000000000000"
	variables["HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_PATH_STYLE"] = "true"
	variables["HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_CREDENTIAL_SOURCE"] = "workload_identity"
	variables["HARPY_IAM_SYSTEM_IMPORT_SOURCE_FCART_KEK_KEYRING_FILE"] =
		"/secrets/iam-system-import-source-fcart-keyring.json"
}
