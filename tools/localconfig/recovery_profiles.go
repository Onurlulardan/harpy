package main

func addRecoveryProfileArtifacts(set *artifactSet) error {
	profiles := map[string]profileEnvironment{
		"database-backup-executor.env":  recoveryBackupEnvironment(),
		"database-restore.env":          recoveryRestoreEnvironment(false),
		"database-rehydrate.env":        recoveryRestoreEnvironment(true),
		"operator-restore-identity.env": recoveryIdentityEnvironment(),
		"operator-restore-license.env":  recoveryLicenseEnvironment(),
	}
	for name, values := range profiles {
		content, err := renderProfileEnvironment(values)
		if err != nil {
			return err
		}
		if err := set.add(name, content); err != nil {
			return err
		}
	}
	return nil
}

func recoveryIdentityEnvironment() profileEnvironment {
	values := baseRuntimeEnvironment(secretRoot + "postgres-runtime-dsn.txt")
	values["HARPY_RESTORE_RECEIPT_PATH"] = "/run/restore-state/receipt.json"
	values["HARPY_RESTORE_SECURITY_EVIDENCE_PATH"] =
		"/run/restore-state/security-evidence.json"
	values["HARPY_IAM_KEY_CUSTODY_MANIFEST_PATH"] =
		"/run/restore-state/iam-key-custody-manifest.json"
	values["HARPY_RECOVERY_TRUST_BUNDLE_PATH"] =
		"/run/recovery-signing/current/recovery-trust-bundle.v1.json"
	values["HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH"] =
		"/run/restore-state/protected-authority-ack-trust-bundle.v1.json"
	return values
}

func recoveryLicenseEnvironment() profileEnvironment {
	values := baseRuntimeEnvironment(secretRoot + "postgres-runtime-dsn.txt")
	values["HARPY_DEPLOYMENT_MODE"] = "onpremise"
	values["HARPY_RESTORE_RECEIPT_PATH"] = "/run/restore-state/receipt.json"
	values["HARPY_RECOVERY_TRUST_BUNDLE_PATH"] =
		"/run/recovery-signing/current/recovery-trust-bundle.v1.json"
	values["HARPY_LICENSE_TRUST_BUNDLE_PATH"] =
		"/run/dev-authority/current/license-trust.json"
	values["HARPY_LICENSE_SNAPSHOT_FILE"] =
		"/run/dev-authority/current/license-snapshot.json"
	return values
}

func recoveryBackupEnvironment() profileEnvironment {
	values := recoveryBaseEnvironment(
		"HARPY_BACKUP_POSTGRES_DSN_FILE",
		secretRoot+"postgres-backup-dsn.txt",
	)
	values["HARPY_RECOVERY_STORAGE_MANIFEST_PATH"] =
		secretRoot + "recovery-storage-manifest.v1.json"
	values["HARPY_RECOVERY_TRUST_BUNDLE_PATH"] =
		"/run/recovery-signing/current/recovery-trust-bundle.v1.json"
	values["HARPY_RECOVERY_MANIFEST_SIGNING_KEY_FILE"] =
		"/run/recovery-signing/current/recovery-manifest-signer.ed25519"
	values["HARPY_MANAGEMENT_LISTEN_ADDR"] = "0.0.0.0:9000"
	return values
}

func recoveryRestoreEnvironment(rehydrate bool) profileEnvironment {
	values := recoveryBaseEnvironment(
		"HARPY_RESTORE_POSTGRES_DSN_FILE",
		secretRoot+"postgres-restore-dsn.txt",
	)
	values["HARPY_RECOVERY_STORAGE_MANIFEST_PATH"] =
		secretRoot + "recovery-storage-manifest.v1.json"
	values["HARPY_RECOVERY_TRUST_BUNDLE_PATH"] =
		"/run/recovery-signing/current/recovery-trust-bundle.v1.json"
	values["HARPY_RESTORE_PROGRESS_PATH"] = "/run/restore-state/progress.json"
	values["HARPY_RESTORE_DRAIN_ACK_PATH"] = "/run/restore-state/drain-ack.json"
	values["HARPY_RESTORE_RECEIPT_PATH"] = "/run/restore-state/receipt.json"
	if rehydrate {
		values["HARPY_REHYDRATE_REQUEST_PATH"] =
			"/run/handoff/rehydrate-request.v1.json"
		values["HARPY_RECOVERY_KIT_FILE"] =
			"/run/handoff/environment-recovery-kit.v1"
		return values
	}
	values["HARPY_RESTORE_PLAN_PATH"] = "/run/handoff/restore-plan.v1.json"
	values["HARPY_DR_KEK_KEYRING_FILE"] =
		"/run/recovery-state/dr-kek-keyring.v1.json"
	values["HARPY_RESTORE_RECEIPT_SIGNING_KEY_FILE"] =
		"/run/recovery-signing/current/restore-receipt-signer.ed25519"
	return values
}

func recoveryBaseEnvironment(dsnKey, dsnPath string) profileEnvironment {
	return profileEnvironment{
		"HARPY_ENVIRONMENT":             "development",
		"HARPY_POSTGRES_ENDPOINT_KIND":  "direct_postgres",
		dsnKey:                          dsnPath,
		"HARPY_RELEASE_MANIFEST_PATH":   releaseRoot + "release.json",
		"HARPY_RELEASE_SIGNATURE_PATH":  releaseRoot + "release.sig",
		"HARPY_RELEASE_TRUST_ROOT_PATH": releaseRoot + "trust-root.json",
	}
}
