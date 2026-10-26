package config

import (
	"os"
	"time"
)

type OperatorInitialize struct {
	runtime              Runtime
	deploymentMode       DeploymentMode
	installationManifest ArtifactPath
	initialPassword      ArtifactPath
	recoveryDirectory    ArtifactPath
	licenseSource        LicenseSource
	licenseTrust         ArtifactPath
	proposerKMSKey       AuthorityKMSKeyID
	registryURL          ServiceURL
	ackTrust             ArtifactPath
	releaseTrust         ArtifactPath
	workloadIdentity     ArtifactPath
	hasWorkloadIdentity  bool
	directoryKeyring     IAMDirectoryKeyring
}

func LoadOperatorInitialize() (OperatorInitialize, error) {
	return loadOperatorInitializeAt(os.Environ(), time.Now().UTC())
}

func loadOperatorInitialize(entries []string) (OperatorInitialize, error) {
	return loadOperatorInitializeAt(entries, time.Now().UTC())
}

func loadOperatorInitializeAt(
	entries []string,
	now time.Time,
) (OperatorInitialize, error) {
	values, runtimeConfig, err := loadRuntime(entries, profileOperatorInitialize)
	if err != nil {
		return OperatorInitialize{}, err
	}
	mode, err := loadDeploymentMode(values)
	if err != nil {
		return OperatorInitialize{}, err
	}
	manifest, err := loadReadArtifact(values, "HARPY_INSTALLATION_MANIFEST_PATH", false)
	if err != nil {
		return OperatorInitialize{}, err
	}
	password, err := loadArtifactPath(values,
		"HARPY_INITIAL_ADMIN_PASSWORD_FILE", ArtifactReadOnceCleanup, true, false)
	if err != nil {
		return OperatorInitialize{}, err
	}
	directory, err := loadIAMHMACKeyring(
		values, iamDirectoryKeyringEnvironmentKey, iamPurposeDirectory, now,
	)
	if err != nil {
		return OperatorInitialize{}, err
	}
	config := OperatorInitialize{
		runtime: runtimeConfig, deploymentMode: mode,
		installationManifest: manifest, initialPassword: password,
		directoryKeyring: IAMDirectoryKeyring{directory},
	}
	if mode == DeploymentSaaS {
		return loadSaaSOperatorInitialize(values, config)
	}
	return loadOnPremOperatorInitialize(values, config)
}

func loadOnPremOperatorInitialize(
	values environmentValues,
	config OperatorInitialize,
) (OperatorInitialize, error) {
	if err := rejectConfigured(values,
		saasKMSKeyIDKey, saasRegistryURLKey, ackTrustBundleKey,
		saasWorkloadIdentityKey, "HARPY_RELEASE_TRUST_ROOT_PATH"); err != nil {
		return OperatorInitialize{}, err
	}
	directory, err := loadArtifactPath(values,
		"HARPY_RECOVERY_SIGNING_DIRECTORY", ArtifactCreateOnly, true, false)
	if err != nil {
		return OperatorInitialize{}, err
	}
	source, err := loadLicenseSource(values, config.runtime.Environment())
	if err != nil {
		return OperatorInitialize{}, err
	}
	trust, err := loadLicenseTrust(values)
	if err != nil {
		return OperatorInitialize{}, err
	}
	config.recoveryDirectory = directory
	config.licenseSource = source
	config.licenseTrust = trust
	return config, nil
}

func loadSaaSOperatorInitialize(
	values environmentValues,
	config OperatorInitialize,
) (OperatorInitialize, error) {
	if err := rejectConfigured(values,
		"HARPY_RECOVERY_SIGNING_DIRECTORY", licenseCredentialKey,
		licenseIssuerURLKey, licenseSnapshotKey, licenseTrustBundleKey); err != nil {
		return OperatorInitialize{}, err
	}
	kmsKey, err := loadAuthorityKMSKeyID(values)
	if err != nil {
		return OperatorInitialize{}, err
	}
	registryURL, err := loadServiceURL(values, saasRegistryURLKey, config.runtime.Environment())
	if err != nil {
		return OperatorInitialize{}, err
	}
	ackTrust, err := loadAuthorityAckTrust(values)
	if err != nil {
		return OperatorInitialize{}, err
	}
	releaseTrust, err := loadReadArtifact(
		values, "HARPY_RELEASE_TRUST_ROOT_PATH", false,
	)
	if err != nil {
		return OperatorInitialize{}, err
	}
	identity, configured, err := loadOptionalReadArtifact(values, saasWorkloadIdentityKey)
	if err != nil {
		return OperatorInitialize{}, err
	}
	config.proposerKMSKey = kmsKey
	config.registryURL = registryURL
	config.ackTrust = ackTrust
	config.releaseTrust = releaseTrust
	config.workloadIdentity = identity
	config.hasWorkloadIdentity = configured
	return config, nil
}
