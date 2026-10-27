package config

const (
	analyticsExportEndpointKey = "HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_ENDPOINT"
	analyticsExportRegionKey   = "HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_REGION"
	analyticsExportBucketKey   = "HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_BUCKET"
	analyticsExportOwnerKey    = "HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_EXPECTED_BUCKET_OWNER"
	analyticsExportPathKey     = "HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_PATH_STYLE"
	analyticsExportSourceKey   = "HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_CREDENTIAL_SOURCE"
	analyticsExportAccessKey   = "HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_ACCESS_KEY_ID_FILE"
	analyticsExportSecretKey   = "HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_SECRET_ACCESS_KEY_FILE"
	analyticsExportSessionKey  = "HARPY_ANALYTICS_DASHBOARD_EXPORT_S3_SESSION_TOKEN_FILE"
	analyticsExportKeyringKey  = "HARPY_ANALYTICS_DASHBOARD_EXPORT_FCART_KEK_KEYRING_FILE"
	analyticsExportSpoolKey    = "HARPY_ANALYTICS_DASHBOARD_EXPORT_SPOOL_DIR"
)

func analyticsDashboardExportStorageKeys() artifactStorageKeys {
	return artifactStorageKeys{
		endpoint: analyticsExportEndpointKey, region: analyticsExportRegionKey,
		bucket: analyticsExportBucketKey, owner: analyticsExportOwnerKey,
		pathStyle: analyticsExportPathKey,
		credentials: artifactCredentialKeys{source: analyticsExportSourceKey,
			access: analyticsExportAccessKey, secret: analyticsExportSecretKey,
			session: analyticsExportSessionKey},
	}
}

func analyticsDashboardExportArtifactKeys() []string {
	return []string{
		analyticsExportEndpointKey, analyticsExportRegionKey, analyticsExportBucketKey,
		analyticsExportOwnerKey, analyticsExportPathKey, analyticsExportSourceKey,
		analyticsExportAccessKey, analyticsExportSecretKey, analyticsExportSessionKey,
		analyticsExportKeyringKey,
	}
}
