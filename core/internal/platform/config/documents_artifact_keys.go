package config

const (
	documentsEndpointKey     = "HARPY_DOCUMENTS_S3_ENDPOINT"
	documentsRegionKey       = "HARPY_DOCUMENTS_S3_REGION"
	documentsBucketKey       = "HARPY_DOCUMENTS_S3_BUCKET"
	documentsOwnerKey        = "HARPY_DOCUMENTS_S3_EXPECTED_BUCKET_OWNER"
	documentsPathKey         = "HARPY_DOCUMENTS_S3_PATH_STYLE"
	documentsSourceKey       = "HARPY_DOCUMENTS_S3_CREDENTIAL_SOURCE"
	documentsAccessKey       = "HARPY_DOCUMENTS_S3_ACCESS_KEY_ID_FILE"
	documentsSecretKey       = "HARPY_DOCUMENTS_S3_SECRET_ACCESS_KEY_FILE"
	documentsSessionKey      = "HARPY_DOCUMENTS_S3_SESSION_TOKEN_FILE"
	documentsBillingActorKey = "HARPY_DOCUMENTS_BILLING_ACTOR_ID"
)

func documentsStorageKeys() artifactStorageKeys {
	return artifactStorageKeys{
		endpoint: documentsEndpointKey, region: documentsRegionKey,
		bucket: documentsBucketKey, owner: documentsOwnerKey, pathStyle: documentsPathKey,
		credentials: artifactCredentialKeys{source: documentsSourceKey,
			access: documentsAccessKey, secret: documentsSecretKey,
			session: documentsSessionKey},
	}
}

func documentsArtifactKeys() []string {
	return []string{
		documentsEndpointKey, documentsRegionKey, documentsBucketKey,
		documentsOwnerKey, documentsPathKey, documentsSourceKey,
		documentsAccessKey, documentsSecretKey, documentsSessionKey,
		documentsKEKKeyringFileKey, documentsBillingActorKey,
	}
}
