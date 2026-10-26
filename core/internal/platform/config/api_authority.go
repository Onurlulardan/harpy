package config

type APIAuthority struct {
	mode                DeploymentMode
	recoveryTrust       ArtifactPath
	recoveryControlKey  ArtifactPath
	ackTrust            ArtifactPath
	proposerKMSKey      AuthorityKMSKeyID
	workloadIdentity    ArtifactPath
	hasWorkloadIdentity bool
}

func loadAPIAuthority(
	values environmentValues,
	mode DeploymentMode,
) (APIAuthority, error) {
	ackTrust, err := loadAuthorityAckTrust(values)
	if err != nil {
		return APIAuthority{}, err
	}
	if mode == DeploymentSaaS {
		if err := rejectConfigured(values,
			recoveryTrustBundleKey, "HARPY_RECOVERY_CONTROL_SIGNING_KEY_FILE"); err != nil {
			return APIAuthority{}, err
		}
		kmsKey, err := loadAuthorityKMSKeyID(values)
		if err != nil {
			return APIAuthority{}, err
		}
		identity, configured, err := loadOptionalReadArtifact(values, saasWorkloadIdentityKey)
		if err != nil {
			return APIAuthority{}, err
		}
		return APIAuthority{
			mode: mode, ackTrust: ackTrust, proposerKMSKey: kmsKey,
			workloadIdentity: identity, hasWorkloadIdentity: configured,
		}, nil
	}
	if err := rejectConfigured(values, saasKMSKeyIDKey, saasWorkloadIdentityKey); err != nil {
		return APIAuthority{}, err
	}
	recoveryTrust, err := loadRecoveryTrust(values)
	if err != nil {
		return APIAuthority{}, err
	}
	controlKey, err := loadReadArtifact(
		values, "HARPY_RECOVERY_CONTROL_SIGNING_KEY_FILE", true,
	)
	if err != nil {
		return APIAuthority{}, err
	}
	return APIAuthority{
		mode: mode, recoveryTrust: recoveryTrust,
		recoveryControlKey: controlKey, ackTrust: ackTrust,
	}, nil
}

func (config APIAuthority) Mode() DeploymentMode                { return config.mode }
func (config APIAuthority) RecoveryTrustBundle() ArtifactPath   { return config.recoveryTrust }
func (config APIAuthority) RecoveryControlKey() ArtifactPath    { return config.recoveryControlKey }
func (config APIAuthority) AckTrustBundle() ArtifactPath        { return config.ackTrust }
func (config APIAuthority) ProposerKMSKeyID() AuthorityKMSKeyID { return config.proposerKMSKey }
func (config APIAuthority) WorkloadIdentity() (ArtifactPath, bool) {
	return config.workloadIdentity, config.hasWorkloadIdentity
}
