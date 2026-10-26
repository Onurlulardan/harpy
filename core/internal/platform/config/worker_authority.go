package config

type WorkerAuthorityMode string

const (
	WorkerAuthorityConnected   WorkerAuthorityMode = "connected"
	WorkerAuthorityAirGap      WorkerAuthorityMode = "airgap"
	WorkerAuthorityDevelopment WorkerAuthorityMode = "development"
	WorkerAuthoritySaaS        WorkerAuthorityMode = "saas"
)

type WorkerAuthority struct {
	mode                WorkerAuthorityMode
	licenseSource       LicenseSource
	licenseTrust        ArtifactPath
	recoveryTrust       ArtifactPath
	ackTrust            ArtifactPath
	releaseTrust        ArtifactPath
	localJournal        ArtifactPath
	localAckSigningKey  ArtifactPath
	registryURL         ServiceURL
	workloadIdentity    ArtifactPath
	hasWorkloadIdentity bool
}

func loadWorkerAuthority(
	values environmentValues,
	environment Environment,
	deploymentMode DeploymentMode,
) (WorkerAuthority, error) {
	ackTrust, err := loadAuthorityAckTrust(values)
	if err != nil {
		return WorkerAuthority{}, err
	}
	if deploymentMode == DeploymentSaaS {
		return loadSaaSWorkerAuthority(values, environment, ackTrust)
	}
	if err := rejectConfigured(values, saasRegistryURLKey, saasWorkloadIdentityKey); err != nil {
		return WorkerAuthority{}, err
	}
	licenseSource, err := loadLicenseSource(values, environment)
	if err != nil {
		return WorkerAuthority{}, err
	}
	licenseTrust, err := loadLicenseTrust(values)
	if err != nil {
		return WorkerAuthority{}, err
	}
	recoveryTrust, err := loadRecoveryTrust(values)
	if err != nil {
		return WorkerAuthority{}, err
	}
	releaseTrust, err := loadReadArtifact(
		values, "HARPY_RELEASE_TRUST_ROOT_PATH", false,
	)
	if err != nil {
		return WorkerAuthority{}, err
	}
	base := WorkerAuthority{
		licenseSource: licenseSource, licenseTrust: licenseTrust,
		recoveryTrust: recoveryTrust, ackTrust: ackTrust, releaseTrust: releaseTrust,
	}
	if environment == EnvironmentDevelopment {
		return loadDevelopmentWorkerAuthority(values, base)
	}
	if err := rejectConfigured(values,
		"HARPY_LOCAL_AUTHORITY_HEAD_JOURNAL_PATH",
		"HARPY_LOCAL_AUTHORITY_ACK_SIGNING_KEY_FILE"); err != nil {
		return WorkerAuthority{}, err
	}
	if licenseSource.Kind() == LicenseSourceConnected {
		base.mode = WorkerAuthorityConnected
	} else {
		base.mode = WorkerAuthorityAirGap
	}
	return base, nil
}

func loadDevelopmentWorkerAuthority(
	values environmentValues,
	base WorkerAuthority,
) (WorkerAuthority, error) {
	if base.licenseSource.Kind() != LicenseSourceSnapshot {
		return WorkerAuthority{}, configError(licenseCredentialKey, "is forbidden in development")
	}
	journal, err := loadArtifactPath(values,
		"HARPY_LOCAL_AUTHORITY_HEAD_JOURNAL_PATH", ArtifactReadWrite, false, true)
	if err != nil {
		return WorkerAuthority{}, err
	}
	signingKey, err := loadReadArtifact(
		values, "HARPY_LOCAL_AUTHORITY_ACK_SIGNING_KEY_FILE", true,
	)
	if err != nil {
		return WorkerAuthority{}, err
	}
	base.mode = WorkerAuthorityDevelopment
	base.localJournal = journal
	base.localAckSigningKey = signingKey
	return base, nil
}

func loadSaaSWorkerAuthority(
	values environmentValues,
	environment Environment,
	ackTrust ArtifactPath,
) (WorkerAuthority, error) {
	if err := rejectConfigured(values,
		licenseCredentialKey, licenseIssuerURLKey, licenseSnapshotKey,
		licenseTrustBundleKey, recoveryTrustBundleKey,
		"HARPY_LOCAL_AUTHORITY_HEAD_JOURNAL_PATH",
		"HARPY_LOCAL_AUTHORITY_ACK_SIGNING_KEY_FILE"); err != nil {
		return WorkerAuthority{}, err
	}
	registryURL, err := loadServiceURL(values, saasRegistryURLKey, environment)
	if err != nil {
		return WorkerAuthority{}, err
	}
	identity, configured, err := loadOptionalReadArtifact(values, saasWorkloadIdentityKey)
	if err != nil {
		return WorkerAuthority{}, err
	}
	releaseTrust, err := loadReadArtifact(
		values, "HARPY_RELEASE_TRUST_ROOT_PATH", false,
	)
	if err != nil {
		return WorkerAuthority{}, err
	}
	return WorkerAuthority{
		mode: WorkerAuthoritySaaS, ackTrust: ackTrust, registryURL: registryURL,
		releaseTrust: releaseTrust, workloadIdentity: identity,
		hasWorkloadIdentity: configured,
	}, nil
}
