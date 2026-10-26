package config

func (config DatabaseRestoreArtifacts) Branch() DatabaseRestoreBranch {
	return config.branch
}
func (config DatabaseRestoreArtifacts) RecoveryStorageManifest() ArtifactPath {
	return config.storageManifest
}
func (config DatabaseRestoreArtifacts) RecoveryTrustBundle() ArtifactPath {
	return config.recoveryTrust
}
func (config DatabaseRestoreArtifacts) OrchestratorTrustBundle() ArtifactPath {
	return config.orchestratorTrust
}
func (config DatabaseRestoreArtifacts) ProgressPath() ArtifactPath {
	return config.progress
}
func (config DatabaseRestoreArtifacts) DrainAckPath() ArtifactPath {
	return config.drainAck
}
func (config DatabaseRestoreArtifacts) SwitchAckPath() ArtifactPath {
	return config.switchAck
}
func (config DatabaseRestoreArtifacts) ReceiptPath() ArtifactPath {
	return config.receipt
}
func (config DatabaseRestoreArtifacts) RestorePlan() ArtifactPath {
	return config.restorePlan
}
func (config DatabaseRestoreArtifacts) DRKEKKeyring() ArtifactPath {
	return config.drKEKKeyring
}
func (config DatabaseRestoreArtifacts) ReceiptSigningKey() ArtifactPath {
	return config.receiptSigningKey
}
func (config DatabaseRestoreArtifacts) RehydrateRequest() ArtifactPath {
	return config.rehydrateRequest
}
func (config DatabaseRestoreArtifacts) RecoveryKit() ArtifactPath {
	return config.recoveryKit
}
