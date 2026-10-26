package config

import "os"

type OperatorIdentityReconcile struct {
	runtime          Runtime
	restoreReceipt   ArtifactPath
	securityEvidence ArtifactPath
	iamCustody       ArtifactPath
	recoveryTrust    ArtifactPath
	ackTrust         ArtifactPath
}

func LoadOperatorIdentityReconcile() (OperatorIdentityReconcile, error) {
	return loadOperatorIdentityReconcile(os.Environ())
}

func loadOperatorIdentityReconcile(entries []string) (OperatorIdentityReconcile, error) {
	values, runtimeConfig, err := loadRuntime(entries, profileOperatorIdentityReconcile)
	if err != nil {
		return OperatorIdentityReconcile{}, err
	}
	receipt, err := loadReadArtifact(values, "HARPY_RESTORE_RECEIPT_PATH", false)
	if err != nil {
		return OperatorIdentityReconcile{}, err
	}
	security, err := loadReadArtifact(values, "HARPY_RESTORE_SECURITY_EVIDENCE_PATH", false)
	if err != nil {
		return OperatorIdentityReconcile{}, err
	}
	custody, err := loadReadArtifact(values, "HARPY_IAM_KEY_CUSTODY_MANIFEST_PATH", false)
	if err != nil {
		return OperatorIdentityReconcile{}, err
	}
	recoveryTrust, err := loadRecoveryTrust(values)
	if err != nil {
		return OperatorIdentityReconcile{}, err
	}
	ackTrust, err := loadAuthorityAckTrust(values)
	if err != nil {
		return OperatorIdentityReconcile{}, err
	}
	if err := validateDistinctArtifactPaths("HARPY_*_PATH",
		receipt, security, custody, recoveryTrust, ackTrust); err != nil {
		return OperatorIdentityReconcile{}, err
	}
	return OperatorIdentityReconcile{
		runtime: runtimeConfig, restoreReceipt: receipt, securityEvidence: security,
		iamCustody: custody, recoveryTrust: recoveryTrust, ackTrust: ackTrust,
	}, nil
}

func (config OperatorIdentityReconcile) Runtime() Runtime { return config.runtime }
func (config OperatorIdentityReconcile) RestoreReceipt() ArtifactPath {
	return config.restoreReceipt
}
func (config OperatorIdentityReconcile) SecurityEvidence() ArtifactPath {
	return config.securityEvidence
}
func (config OperatorIdentityReconcile) IAMCustodyManifest() ArtifactPath {
	return config.iamCustody
}
func (config OperatorIdentityReconcile) RecoveryTrustBundle() ArtifactPath {
	return config.recoveryTrust
}
func (config OperatorIdentityReconcile) AckTrustBundle() ArtifactPath {
	return config.ackTrust
}
