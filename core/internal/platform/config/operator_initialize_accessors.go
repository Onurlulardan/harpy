package config

func (config OperatorInitialize) Runtime() Runtime { return config.runtime }
func (config OperatorInitialize) DeploymentMode() DeploymentMode {
	return config.deploymentMode
}
func (config OperatorInitialize) InstallationManifest() ArtifactPath {
	return config.installationManifest
}
func (config OperatorInitialize) InitialPassword() ArtifactPath {
	return config.initialPassword
}
func (config OperatorInitialize) DirectoryKeyring() (IAMDirectoryKeyring, error) {
	if err := config.directoryKeyring.Validate(); err != nil ||
		config.directoryKeyring.purpose != iamPurposeDirectory {
		return IAMDirectoryKeyring{}, errInvalidIAMHMACKeyring
	}
	return config.directoryKeyring, nil
}
func (config OperatorInitialize) RecoverySigningDirectory() ArtifactPath {
	return config.recoveryDirectory
}
func (config OperatorInitialize) LicenseSource() LicenseSource {
	return config.licenseSource
}
func (config OperatorInitialize) LicenseTrustBundle() ArtifactPath {
	return config.licenseTrust
}
func (config OperatorInitialize) ProposerKMSKeyID() AuthorityKMSKeyID {
	return config.proposerKMSKey
}
func (config OperatorInitialize) AuthorityRegistryURL() ServiceURL {
	return config.registryURL
}
func (config OperatorInitialize) AckTrustBundle() ArtifactPath   { return config.ackTrust }
func (config OperatorInitialize) ReleaseTrustRoot() ArtifactPath { return config.releaseTrust }
func (config OperatorInitialize) WorkloadIdentity() (ArtifactPath, bool) {
	return config.workloadIdentity, config.hasWorkloadIdentity
}
