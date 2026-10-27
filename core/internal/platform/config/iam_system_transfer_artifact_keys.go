package config

const (
	iamTransferSpoolKey = "HARPY_IAM_SYSTEM_TRANSFER_SPOOL_DIR"

	iamExportEndpointKey = "HARPY_IAM_SYSTEM_EXPORT_S3_ENDPOINT"
	iamExportRegionKey   = "HARPY_IAM_SYSTEM_EXPORT_S3_REGION"
	iamExportBucketKey   = "HARPY_IAM_SYSTEM_EXPORT_S3_BUCKET"
	iamExportOwnerKey    = "HARPY_IAM_SYSTEM_EXPORT_S3_EXPECTED_BUCKET_OWNER"
	iamExportPathKey     = "HARPY_IAM_SYSTEM_EXPORT_S3_PATH_STYLE"
	iamExportSourceKey   = "HARPY_IAM_SYSTEM_EXPORT_S3_CREDENTIAL_SOURCE"
	iamExportAccessKey   = "HARPY_IAM_SYSTEM_EXPORT_S3_ACCESS_KEY_ID_FILE"
	iamExportSecretKey   = "HARPY_IAM_SYSTEM_EXPORT_S3_SECRET_ACCESS_KEY_FILE"
	iamExportSessionKey  = "HARPY_IAM_SYSTEM_EXPORT_S3_SESSION_TOKEN_FILE"
	iamExportKeyringKey  = "HARPY_IAM_SYSTEM_EXPORT_FCART_KEK_KEYRING_FILE"

	iamImportEndpointKey = "HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_ENDPOINT"
	iamImportRegionKey   = "HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_REGION"
	iamImportBucketKey   = "HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_BUCKET"
	iamImportOwnerKey    = "HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_EXPECTED_BUCKET_OWNER"
	iamImportPathKey     = "HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_PATH_STYLE"
	iamImportSourceKey   = "HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_CREDENTIAL_SOURCE"
	iamImportAccessKey   = "HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_ACCESS_KEY_ID_FILE"
	iamImportSecretKey   = "HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_SECRET_ACCESS_KEY_FILE"
	iamImportSessionKey  = "HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_SESSION_TOKEN_FILE"
	iamImportKeyringKey  = "HARPY_IAM_SYSTEM_IMPORT_SOURCE_FCART_KEK_KEYRING_FILE"
)

func iamExportStorageKeys() artifactStorageKeys {
	return artifactStorageKeys{endpoint: iamExportEndpointKey, region: iamExportRegionKey,
		bucket: iamExportBucketKey, owner: iamExportOwnerKey, pathStyle: iamExportPathKey,
		credentials: artifactCredentialKeys{source: iamExportSourceKey,
			access: iamExportAccessKey, secret: iamExportSecretKey,
			session: iamExportSessionKey}}
}

func iamImportStorageKeys() artifactStorageKeys {
	return artifactStorageKeys{endpoint: iamImportEndpointKey, region: iamImportRegionKey,
		bucket: iamImportBucketKey, owner: iamImportOwnerKey, pathStyle: iamImportPathKey,
		credentials: artifactCredentialKeys{source: iamImportSourceKey,
			access: iamImportAccessKey, secret: iamImportSecretKey,
			session: iamImportSessionKey}}
}

func iamSystemTransferArtifactKeys() []string {
	return []string{iamTransferSpoolKey,
		iamExportEndpointKey, iamExportRegionKey, iamExportBucketKey,
		iamExportOwnerKey, iamExportPathKey, iamExportSourceKey,
		iamExportAccessKey, iamExportSecretKey, iamExportSessionKey, iamExportKeyringKey,
		iamImportEndpointKey, iamImportRegionKey, iamImportBucketKey,
		iamImportOwnerKey, iamImportPathKey, iamImportSourceKey,
		iamImportAccessKey, iamImportSecretKey, iamImportSessionKey, iamImportKeyringKey}
}
