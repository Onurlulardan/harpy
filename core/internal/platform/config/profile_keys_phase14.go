package config

const (
	ackTrustBundleKey       = "HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH"
	recoveryTrustBundleKey  = "HARPY_RECOVERY_TRUST_BUNDLE_PATH"
	licenseTrustBundleKey   = "HARPY_LICENSE_TRUST_BUNDLE_PATH"
	licenseCredentialKey    = "HARPY_LICENSE_CREDENTIAL_FILE"
	licenseIssuerURLKey     = "HARPY_LICENSE_ISSUER_URL"
	licenseSnapshotKey      = "HARPY_LICENSE_SNAPSHOT_FILE"
	saasKMSKeyIDKey         = "HARPY_SAAS_AUTHORITY_PROPOSER_KMS_KEY_ID"
	saasRegistryURLKey      = "HARPY_SAAS_AUTHORITY_REGISTRY_URL"
	saasWorkloadIdentityKey = "HARPY_SAAS_AUTHORITY_WORKLOAD_IDENTITY_CONFIG_PATH"
)

func apiPhase14Keys() []string {
	return []string{
		recoveryTrustBundleKey, "HARPY_RECOVERY_CONTROL_SIGNING_KEY_FILE",
		ackTrustBundleKey, saasKMSKeyIDKey, saasWorkloadIdentityKey,
	}
}

func workerPhase14Keys() []string {
	return []string{
		licenseCredentialKey, licenseIssuerURLKey, licenseSnapshotKey,
		licenseTrustBundleKey, recoveryTrustBundleKey, ackTrustBundleKey,
		"HARPY_LOCAL_AUTHORITY_HEAD_JOURNAL_PATH",
		"HARPY_LOCAL_AUTHORITY_ACK_SIGNING_KEY_FILE",
		saasRegistryURLKey, saasWorkloadIdentityKey,
		"HARPY_RELEASE_TRUST_ROOT_PATH",
	}
}

func operatorInitializeKeys() []string {
	return []string{
		deploymentModeEnvironmentKey, "HARPY_INSTALLATION_MANIFEST_PATH",
		"HARPY_INITIAL_ADMIN_PASSWORD_FILE", "HARPY_RECOVERY_SIGNING_DIRECTORY",
		iamDirectoryKeyringEnvironmentKey, iamDirectoryKeyringEnvironmentKey + "_FILE",
		licenseCredentialKey, licenseIssuerURLKey, licenseSnapshotKey,
		licenseTrustBundleKey, saasKMSKeyIDKey, saasRegistryURLKey,
		ackTrustBundleKey, saasWorkloadIdentityKey, "HARPY_RELEASE_TRUST_ROOT_PATH",
	}
}

func operatorIdentityReconcileKeys() []string {
	return []string{
		"HARPY_RESTORE_RECEIPT_PATH", "HARPY_RESTORE_SECURITY_EVIDENCE_PATH",
		"HARPY_IAM_KEY_CUSTODY_MANIFEST_PATH", recoveryTrustBundleKey,
		ackTrustBundleKey,
	}
}

func operatorLicenseReconcileKeys() []string {
	return []string{
		deploymentModeEnvironmentKey, "HARPY_RESTORE_RECEIPT_PATH",
		recoveryTrustBundleKey, licenseTrustBundleKey, licenseCredentialKey,
		licenseIssuerURLKey, licenseSnapshotKey,
	}
}

func postRestoreObserverKeys() []string {
	keys := dedicatedProfileKeys(
		"HARPY_POST_RESTORE_REACTIVATION_POSTGRES_DSN_FILE", false,
	)
	return append(
		keys, "HARPY_POST_RESTORE_REACTIVATION_INPUT_FILE",
	)
}

func releaseInputKeys() []string {
	return []string{
		"HARPY_RELEASE_MANIFEST_PATH", "HARPY_RELEASE_SIGNATURE_PATH",
		"HARPY_RELEASE_TRUST_ROOT_PATH",
	}
}

func dedicatedProfileKeys(dsnFile string, longRunning bool) []string {
	keys := []string{
		"HARPY_ENVIRONMENT", "HARPY_POSTGRES_ENDPOINT_KIND", dsnFile,
		"HARPY_POSTGRES_CONNECT_TIMEOUT", "HARPY_POSTGRES_CLEANUP_TIMEOUT",
		"HARPY_LOG_LEVEL", "HARPY_OTLP_ENDPOINT", "HARPY_OTLP_HEADERS",
		"HARPY_OTLP_HEADERS_FILE", "HARPY_OTLP_INSECURE",
		"HARPY_OTLP_EXPORT_TIMEOUT", "HARPY_TRACE_SAMPLE_RATIO",
	}
	keys = append(keys, releaseInputKeys()...)
	if longRunning {
		keys = append(keys,
			"HARPY_POSTGRES_MAX_CONNS", "HARPY_POSTGRES_MIN_CONNS",
			"HARPY_POSTGRES_MAX_CONN_LIFETIME", "HARPY_POSTGRES_MAX_CONN_IDLE_TIME",
			"HARPY_POSTGRES_HEALTH_CHECK_PERIOD", "HARPY_READINESS_TIMEOUT",
			"HARPY_MANAGEMENT_LISTEN_ADDR", "HARPY_DRAIN_TIMEOUT")
	}
	return keys
}

func databaseBackupExecutorKeys() []string {
	keys := dedicatedProfileKeys("HARPY_BACKUP_POSTGRES_DSN_FILE", true)
	return append(keys, "HARPY_RECOVERY_STORAGE_MANIFEST_PATH",
		recoveryTrustBundleKey, "HARPY_RECOVERY_MANIFEST_SIGNING_KEY_FILE")
}

func databaseRestoreKeys() []string {
	keys := dedicatedProfileKeys(
		"HARPY_RESTORE_CONTROL_POSTGRES_DSN_FILE", false,
	)
	keys = append(keys, "HARPY_RESTORE_SCRATCH_POSTGRES_DSN_FILE")
	return append(keys,
		"HARPY_RECOVERY_STORAGE_MANIFEST_PATH", recoveryTrustBundleKey,
		"HARPY_RESTORE_ORCHESTRATOR_TRUST_BUNDLE_PATH",
		"HARPY_RESTORE_PROGRESS_PATH", "HARPY_RESTORE_DRAIN_ACK_PATH",
		"HARPY_RESTORE_SWITCH_ACK_PATH",
		"HARPY_RESTORE_RECEIPT_PATH", "HARPY_RESTORE_PLAN_PATH",
		"HARPY_DR_KEK_KEYRING_FILE", "HARPY_RESTORE_RECEIPT_SIGNING_KEY_FILE",
		"HARPY_REHYDRATE_REQUEST_PATH", "HARPY_RECOVERY_KIT_FILE")
}
