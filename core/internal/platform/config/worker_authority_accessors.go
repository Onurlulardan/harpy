package config

func (config WorkerAuthority) Mode() WorkerAuthorityMode { return config.mode }
func (config WorkerAuthority) LicenseSource() LicenseSource {
	return config.licenseSource
}
func (config WorkerAuthority) LicenseTrustBundle() ArtifactPath {
	return config.licenseTrust
}
func (config WorkerAuthority) RecoveryTrustBundle() ArtifactPath {
	return config.recoveryTrust
}
func (config WorkerAuthority) AckTrustBundle() ArtifactPath   { return config.ackTrust }
func (config WorkerAuthority) ReleaseTrustRoot() ArtifactPath { return config.releaseTrust }
func (config WorkerAuthority) LocalJournal() ArtifactPath     { return config.localJournal }
func (config WorkerAuthority) LocalAckSigningKey() ArtifactPath {
	return config.localAckSigningKey
}
func (config WorkerAuthority) RegistryURL() ServiceURL { return config.registryURL }
func (config WorkerAuthority) WorkloadIdentity() (ArtifactPath, bool) {
	return config.workloadIdentity, config.hasWorkloadIdentity
}
