package config

import "os"

type OperatorLicenseReconcile struct {
	runtime        Runtime
	deploymentMode DeploymentMode
	restoreReceipt ArtifactPath
	recoveryTrust  ArtifactPath
	licenseTrust   ArtifactPath
	licenseSource  LicenseSource
}

func LoadOperatorLicenseReconcile() (OperatorLicenseReconcile, error) {
	return loadOperatorLicenseReconcile(os.Environ())
}

func loadOperatorLicenseReconcile(entries []string) (OperatorLicenseReconcile, error) {
	values, runtimeConfig, err := loadRuntime(entries, profileOperatorLicenseReconcile)
	if err != nil {
		return OperatorLicenseReconcile{}, err
	}
	mode, err := loadDeploymentMode(values)
	if err != nil {
		return OperatorLicenseReconcile{}, err
	}
	if mode != DeploymentOnPrem {
		return OperatorLicenseReconcile{}, configError(
			deploymentModeEnvironmentKey, "license reconcile is onpremise-only",
		)
	}
	receipt, err := loadReadArtifact(values, "HARPY_RESTORE_RECEIPT_PATH", false)
	if err != nil {
		return OperatorLicenseReconcile{}, err
	}
	recoveryTrust, err := loadRecoveryTrust(values)
	if err != nil {
		return OperatorLicenseReconcile{}, err
	}
	licenseTrust, err := loadLicenseTrust(values)
	if err != nil {
		return OperatorLicenseReconcile{}, err
	}
	source, err := loadLicenseSource(values, runtimeConfig.Environment())
	if err != nil {
		return OperatorLicenseReconcile{}, err
	}
	paths := []ArtifactPath{receipt, recoveryTrust, licenseTrust}
	if source.Kind() == LicenseSourceConnected {
		paths = append(paths, source.CredentialFile())
	} else {
		paths = append(paths, source.SnapshotFile())
	}
	if err := validateDistinctArtifactPaths("HARPY_*_PATH", paths...); err != nil {
		return OperatorLicenseReconcile{}, err
	}
	return OperatorLicenseReconcile{
		runtime: runtimeConfig, deploymentMode: mode, restoreReceipt: receipt,
		recoveryTrust: recoveryTrust, licenseTrust: licenseTrust, licenseSource: source,
	}, nil
}

func (config OperatorLicenseReconcile) Runtime() Runtime { return config.runtime }
func (config OperatorLicenseReconcile) DeploymentMode() DeploymentMode {
	return config.deploymentMode
}
func (config OperatorLicenseReconcile) RestoreReceipt() ArtifactPath {
	return config.restoreReceipt
}
func (config OperatorLicenseReconcile) RecoveryTrustBundle() ArtifactPath {
	return config.recoveryTrust
}
func (config OperatorLicenseReconcile) LicenseTrustBundle() ArtifactPath {
	return config.licenseTrust
}
func (config OperatorLicenseReconcile) LicenseSource() LicenseSource {
	return config.licenseSource
}
