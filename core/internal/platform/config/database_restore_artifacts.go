package config

type DatabaseRestoreBranch string

const (
	RestoreExistingReadable DatabaseRestoreBranch = "existing_readable"
	RestoreRehydrate        DatabaseRestoreBranch = "empty_or_unreadable"
)

type DatabaseRestoreArtifacts struct {
	branch            DatabaseRestoreBranch
	storageManifest   ArtifactPath
	recoveryTrust     ArtifactPath
	orchestratorTrust ArtifactPath
	progress          ArtifactPath
	drainAck          ArtifactPath
	switchAck         ArtifactPath
	receipt           ArtifactPath
	restorePlan       ArtifactPath
	drKEKKeyring      ArtifactPath
	receiptSigningKey ArtifactPath
	rehydrateRequest  ArtifactPath
	recoveryKit       ArtifactPath
}

func loadDatabaseRestoreArtifacts(
	values environmentValues,
	dsnFiles ...ArtifactPath,
) (DatabaseRestoreArtifacts, error) {
	storage, err := loadReadArtifact(values, "HARPY_RECOVERY_STORAGE_MANIFEST_PATH", false)
	if err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	orchestratorTrust, err := loadReadArtifact(
		values, "HARPY_RESTORE_ORCHESTRATOR_TRUST_BUNDLE_PATH", false,
	)
	if err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	trust, err := loadRecoveryTrust(values)
	if err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	switchAck, err := loadArtifactPath(values,
		"HARPY_RESTORE_SWITCH_ACK_PATH", ArtifactReadOnly, false, false)
	if err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	progress, err := loadArtifactPath(values,
		"HARPY_RESTORE_PROGRESS_PATH", ArtifactAtomicWrite, false, false)
	if err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	drainAck, err := loadArtifactPath(values,
		"HARPY_RESTORE_DRAIN_ACK_PATH", ArtifactReadOnly, false, false)
	if err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	receipt, err := loadArtifactPath(values,
		"HARPY_RESTORE_RECEIPT_PATH", ArtifactAtomicWrite, false, false)
	if err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	config := DatabaseRestoreArtifacts{
		storageManifest: storage, recoveryTrust: trust,
		orchestratorTrust: orchestratorTrust, progress: progress,
		drainAck: drainAck, switchAck: switchAck, receipt: receipt,
	}
	existing := anyConfigured(values,
		"HARPY_RESTORE_PLAN_PATH", "HARPY_DR_KEK_KEYRING_FILE",
		"HARPY_RESTORE_RECEIPT_SIGNING_KEY_FILE")
	rehydrate := anyConfigured(values,
		"HARPY_REHYDRATE_REQUEST_PATH", "HARPY_RECOVERY_KIT_FILE")
	if existing == rehydrate {
		return DatabaseRestoreArtifacts{}, configError(
			"HARPY_RESTORE_*", "restore branches must form an exact XOR")
	}
	if existing {
		config, err = loadExistingRestoreBranch(values, config)
	} else {
		config, err = loadRehydrateRestoreBranch(values, config)
	}
	if err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	paths := append(append([]ArtifactPath(nil), dsnFiles...), config.paths()...)
	if err := validateDistinctArtifactPaths("HARPY_RESTORE_*_PATH", paths...); err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	return config, nil
}

func loadExistingRestoreBranch(
	values environmentValues,
	config DatabaseRestoreArtifacts,
) (DatabaseRestoreArtifacts, error) {
	plan, err := loadReadArtifact(values, "HARPY_RESTORE_PLAN_PATH", false)
	if err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	keyring, err := loadReadArtifact(values, "HARPY_DR_KEK_KEYRING_FILE", true)
	if err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	signingKey, err := loadReadArtifact(
		values, "HARPY_RESTORE_RECEIPT_SIGNING_KEY_FILE", true)
	if err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	config.branch = RestoreExistingReadable
	config.restorePlan = plan
	config.drKEKKeyring = keyring
	config.receiptSigningKey = signingKey
	return config, nil
}

func loadRehydrateRestoreBranch(
	values environmentValues,
	config DatabaseRestoreArtifacts,
) (DatabaseRestoreArtifacts, error) {
	request, err := loadReadArtifact(values, "HARPY_REHYDRATE_REQUEST_PATH", false)
	if err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	kit, err := loadArtifactPath(values,
		"HARPY_RECOVERY_KIT_FILE", ArtifactReadOnceCleanup, true, true)
	if err != nil {
		return DatabaseRestoreArtifacts{}, err
	}
	config.branch = RestoreRehydrate
	config.rehydrateRequest = request
	config.recoveryKit = kit
	return config, nil
}

func anyConfigured(values environmentValues, keys ...string) bool {
	for _, key := range keys {
		if _, found := values.lookup(key); found {
			return true
		}
	}
	return false
}

func (config DatabaseRestoreArtifacts) paths() []ArtifactPath {
	paths := []ArtifactPath{
		config.storageManifest, config.recoveryTrust, config.orchestratorTrust,
		config.progress, config.drainAck, config.switchAck, config.receipt,
	}
	if config.branch == RestoreExistingReadable {
		return append(paths, config.restorePlan, config.drKEKKeyring, config.receiptSigningKey)
	}
	return append(paths, config.rehydrateRequest, config.recoveryKit)
}
