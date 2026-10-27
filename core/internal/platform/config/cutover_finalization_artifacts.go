package config

type CutoverFinalizationArtifacts struct {
	leaseManifestPath           AbsolutePath
	leaseSignaturePath          AbsolutePath
	trustRootPath               AbsolutePath
	sessionSealReceiptPath      AbsolutePath
	credentialCleanupPath       AbsolutePath
	finalizerLeaseManifestPath  AbsolutePath
	finalizerLeaseSignaturePath AbsolutePath
	releaseManifestPath         AbsolutePath
	releaseSignaturePath        AbsolutePath
	bootstrapAuthorityPath      AbsolutePath
	bootstrapAuthoritySignature AbsolutePath
	authorityOperationPath      AbsolutePath
	authorityOperationSignature AbsolutePath
	authorityReceiptPath        AbsolutePath
	authorityReceiptSignature   AbsolutePath
	terminalEvidencePath        AbsolutePath
}

func loadCutoverFinalizationArtifacts(
	values environmentValues,
) (CutoverFinalizationArtifacts, error) {
	keys := []string{
		"HARPY_CUTOVER_LEASE_MANIFEST_PATH",
		"HARPY_CUTOVER_LEASE_SIGNATURE_PATH",
		"HARPY_CUTOVER_TRUST_ROOT_PATH",
		"HARPY_CUTOVER_SESSION_SEAL_RECEIPT_PATH",
		"HARPY_CUTOVER_CREDENTIAL_CLEANUP_RECEIPT_PATH",
		"HARPY_CUTOVER_FINALIZER_LEASE_MANIFEST_PATH",
		"HARPY_CUTOVER_FINALIZER_LEASE_SIGNATURE_PATH",
		"HARPY_RELEASE_MANIFEST_PATH",
		"HARPY_RELEASE_SIGNATURE_PATH",
		"HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH",
		"HARPY_BOOTSTRAP_AUTHORITY_SIGNATURE_PATH",
		"HARPY_CUTOVER_AUTHORITY_OPERATION_PATH",
		"HARPY_CUTOVER_AUTHORITY_OPERATION_SIGNATURE_PATH",
		"HARPY_CUTOVER_AUTHORITY_RECEIPT_PATH",
		"HARPY_CUTOVER_AUTHORITY_RECEIPT_SIGNATURE_PATH",
		"HARPY_CUTOVER_TERMINAL_EVIDENCE_PATH",
	}
	paths, err := loadDistinctCutoverPaths(values, keys)
	if err != nil {
		return CutoverFinalizationArtifacts{}, err
	}
	return CutoverFinalizationArtifacts{
		leaseManifestPath: paths[0], leaseSignaturePath: paths[1], trustRootPath: paths[2],
		sessionSealReceiptPath: paths[3], credentialCleanupPath: paths[4],
		finalizerLeaseManifestPath: paths[5], finalizerLeaseSignaturePath: paths[6],
		releaseManifestPath: paths[7], releaseSignaturePath: paths[8],
		bootstrapAuthorityPath: paths[9], bootstrapAuthoritySignature: paths[10],
		authorityOperationPath: paths[11], authorityOperationSignature: paths[12],
		authorityReceiptPath: paths[13], authorityReceiptSignature: paths[14],
		terminalEvidencePath: paths[15],
	}, nil
}

func (a CutoverFinalizationArtifacts) LeaseManifestPath() AbsolutePath  { return a.leaseManifestPath }
func (a CutoverFinalizationArtifacts) LeaseSignaturePath() AbsolutePath { return a.leaseSignaturePath }
func (a CutoverFinalizationArtifacts) TrustRootPath() AbsolutePath      { return a.trustRootPath }
func (a CutoverFinalizationArtifacts) SessionSealReceiptPath() AbsolutePath {
	return a.sessionSealReceiptPath
}
func (a CutoverFinalizationArtifacts) CredentialCleanupPath() AbsolutePath {
	return a.credentialCleanupPath
}
func (a CutoverFinalizationArtifacts) FinalizerLeaseManifestPath() AbsolutePath {
	return a.finalizerLeaseManifestPath
}
func (a CutoverFinalizationArtifacts) FinalizerLeaseSignaturePath() AbsolutePath {
	return a.finalizerLeaseSignaturePath
}
func (a CutoverFinalizationArtifacts) ReleaseManifestPath() AbsolutePath {
	return a.releaseManifestPath
}
func (a CutoverFinalizationArtifacts) ReleaseSignaturePath() AbsolutePath {
	return a.releaseSignaturePath
}
func (a CutoverFinalizationArtifacts) BootstrapAuthorityPath() AbsolutePath {
	return a.bootstrapAuthorityPath
}
func (a CutoverFinalizationArtifacts) BootstrapAuthoritySignaturePath() AbsolutePath {
	return a.bootstrapAuthoritySignature
}
func (a CutoverFinalizationArtifacts) AuthorityOperationPath() AbsolutePath {
	return a.authorityOperationPath
}
func (a CutoverFinalizationArtifacts) AuthorityOperationSignaturePath() AbsolutePath {
	return a.authorityOperationSignature
}
func (a CutoverFinalizationArtifacts) AuthorityReceiptPath() AbsolutePath {
	return a.authorityReceiptPath
}
func (a CutoverFinalizationArtifacts) AuthorityReceiptSignaturePath() AbsolutePath {
	return a.authorityReceiptSignature
}
func (a CutoverFinalizationArtifacts) TerminalEvidencePath() AbsolutePath {
	return a.terminalEvidencePath
}
