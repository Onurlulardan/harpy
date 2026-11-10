package config

func loadRecoveryTrust(values environmentValues) (ArtifactPath, error) {
	return loadReadArtifact(values, recoveryTrustBundleKey, false)
}

func loadAuthorityAckTrust(values environmentValues) (ArtifactPath, error) {
	return loadReadArtifact(values, ackTrustBundleKey, false)
}
