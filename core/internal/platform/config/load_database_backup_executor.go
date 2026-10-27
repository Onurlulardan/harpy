package config

import "os"

type DatabaseBackupExecutor struct {
	longRunning     DedicatedLongRunning
	storageManifest ArtifactPath
	recoveryTrust   ArtifactPath
	manifestSigning ArtifactPath
}

func LoadDatabaseBackupExecutor() (DatabaseBackupExecutor, error) {
	return loadDatabaseBackupExecutor(os.Environ())
}

func loadDatabaseBackupExecutor(entries []string) (DatabaseBackupExecutor, error) {
	values, runtimeConfig, err := loadDedicatedRuntime(entries,
		profileDatabaseBackupExecutor, "HARPY_BACKUP_POSTGRES_DSN_FILE", "harpy_backup")
	if err != nil {
		return DatabaseBackupExecutor{}, err
	}
	longRunning, err := loadDedicatedLongRunning(values, runtimeConfig)
	if err != nil {
		return DatabaseBackupExecutor{}, err
	}
	storage, err := loadReadArtifact(values, "HARPY_RECOVERY_STORAGE_MANIFEST_PATH", false)
	if err != nil {
		return DatabaseBackupExecutor{}, err
	}
	trust, err := loadRecoveryTrust(values)
	if err != nil {
		return DatabaseBackupExecutor{}, err
	}
	signingKey, err := loadReadArtifact(
		values, "HARPY_RECOVERY_MANIFEST_SIGNING_KEY_FILE", true,
	)
	if err != nil {
		return DatabaseBackupExecutor{}, err
	}
	if err := validateDistinctArtifactPaths("HARPY_*_PATH",
		runtimeConfig.DSNFile(), storage, trust, signingKey); err != nil {
		return DatabaseBackupExecutor{}, err
	}
	return DatabaseBackupExecutor{
		longRunning: longRunning, storageManifest: storage,
		recoveryTrust: trust, manifestSigning: signingKey,
	}, nil
}

func (config DatabaseBackupExecutor) LongRunning() DedicatedLongRunning {
	return config.longRunning
}
func (config DatabaseBackupExecutor) RecoveryStorageManifest() ArtifactPath {
	return config.storageManifest
}
func (config DatabaseBackupExecutor) RecoveryTrustBundle() ArtifactPath {
	return config.recoveryTrust
}
func (config DatabaseBackupExecutor) ManifestSigningKey() ArtifactPath {
	return config.manifestSigning
}
