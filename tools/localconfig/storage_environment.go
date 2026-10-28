package main

const (
	secretRoot = "/run/secrets/"
	localS3URL = "http://s3-owner:9000"
	localOwner = "123456789012"
)

type storageSurface struct {
	prefix string
	bucket string
}

var sharedStorageSurfaces = []storageSurface{
	{"HARPY_DOCUMENTS", "harpy-documents"},
	{"HARPY_ANALYTICS_DASHBOARD_EXPORT", "harpy-analytics-dashboard-exports"},
	{"HARPY_IAM_SYSTEM_EXPORT", "harpy-iam-system-exports"},
	{"HARPY_IAM_SYSTEM_IMPORT_SOURCE", "harpy-iam-system-import-sources"},
}

func addSharedStorageEnvironment(values profileEnvironment) {
	for _, surface := range sharedStorageSurfaces {
		addS3Environment(values, surface.prefix, surface.bucket)
	}
	values["HARPY_DOCUMENTS_KEK_KEYRING_FILE"] =
		secretRoot + "documents-kek-keyring.json"
	values["HARPY_DOCUMENTS_BILLING_ACTOR_ID"] =
		"019b1234-5678-7abc-8def-0123456789ab"
	values["HARPY_ANALYTICS_DASHBOARD_EXPORT_FCART_KEK_KEYRING_FILE"] =
		secretRoot + "analytics-dashboard-export-fcart-kek-keyring.json"
	values["HARPY_IAM_SYSTEM_EXPORT_FCART_KEK_KEYRING_FILE"] =
		secretRoot + "iam-system-export-fcart-kek-keyring.json"
	values["HARPY_IAM_SYSTEM_IMPORT_SOURCE_FCART_KEK_KEYRING_FILE"] =
		secretRoot + "iam-system-import-source-fcart-kek-keyring.json"
	values["HARPY_IAM_SYSTEM_TRANSFER_SPOOL_DIR"] =
		"/var/lib/harpy/iam-system-transfer"
}

func addS3Environment(
	values profileEnvironment,
	prefix, bucket string,
) {
	key := prefix + "_S3_"
	values[key+"ENDPOINT"] = localS3URL
	values[key+"REGION"] = "us-east-1"
	values[key+"BUCKET"] = bucket
	values[key+"EXPECTED_BUCKET_OWNER"] = localOwner
	values[key+"PATH_STYLE"] = "true"
	values[key+"CREDENTIAL_SOURCE"] = "static_files"
	values[key+"ACCESS_KEY_ID_FILE"] = secretRoot + "s3-access-key-id.txt"
	values[key+"SECRET_ACCESS_KEY_FILE"] = secretRoot + "s3-secret-access-key.txt"
}

func addSharedKeyringEnvironment(values profileEnvironment) {
	values["HARPY_REST_QUERY_CURSOR_KEYRING_FILE"] =
		secretRoot + "rest-query-cursor-keyring.json"
	values["HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING_FILE"] =
		secretRoot + "catalog-package-public-keyring.json"
	values["HARPY_IAM_LOGIN_DIRECTORY_KEYRING_FILE"] =
		secretRoot + "iam-login-directory-keyring.json"
	values["HARPY_IAM_ONE_TIME_TOKEN_KEYRING_FILE"] =
		secretRoot + "iam-one-time-token-keyring.json"
	values["HARPY_IAM_DELIVERY_KMS_KEY_ID"] = localDeliveryKMSKeyID
}
