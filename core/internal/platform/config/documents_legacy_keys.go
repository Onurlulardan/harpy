package config

const (
	documentsLegacyEndpointKey = "HARPY_DOCUMENTS_LEGACY_SOURCE_S3_ENDPOINT"
	documentsLegacyRegionKey   = "HARPY_DOCUMENTS_LEGACY_SOURCE_S3_REGION"
	documentsLegacyBucketKey   = "HARPY_DOCUMENTS_LEGACY_SOURCE_S3_BUCKET"
	documentsLegacyOwnerKey    = "HARPY_DOCUMENTS_LEGACY_SOURCE_S3_EXPECTED_BUCKET_OWNER"
	documentsLegacyPathKey     = "HARPY_DOCUMENTS_LEGACY_SOURCE_S3_PATH_STYLE"
	documentsLegacySourceKey   = "HARPY_DOCUMENTS_LEGACY_SOURCE_S3_CREDENTIAL_SOURCE"
	documentsLegacyAccessKey   = "HARPY_DOCUMENTS_LEGACY_SOURCE_S3_ACCESS_KEY_ID_FILE"
	documentsLegacySecretKey   = "HARPY_DOCUMENTS_LEGACY_SOURCE_S3_SECRET_ACCESS_KEY_FILE"
	documentsLegacySessionKey  = "HARPY_DOCUMENTS_LEGACY_SOURCE_S3_SESSION_TOKEN_FILE"
)

func documentsLegacyStorageKeys() artifactStorageKeys {
	return artifactStorageKeys{
		endpoint: documentsLegacyEndpointKey, region: documentsLegacyRegionKey,
		bucket: documentsLegacyBucketKey, owner: documentsLegacyOwnerKey,
		pathStyle: documentsLegacyPathKey,
		credentials: artifactCredentialKeys{source: documentsLegacySourceKey,
			access: documentsLegacyAccessKey, secret: documentsLegacySecretKey,
			session: documentsLegacySessionKey},
	}
}

func documentsLegacyArtifactKeys() []string {
	return []string{
		documentsLegacyEndpointKey, documentsLegacyRegionKey, documentsLegacyBucketKey,
		documentsLegacyOwnerKey, documentsLegacyPathKey, documentsLegacySourceKey,
		documentsLegacyAccessKey, documentsLegacySecretKey, documentsLegacySessionKey,
		documentsArchiveReservationKey,
	}
}
