package artifact

func validCheckpoint(value Checkpoint) bool {
	_, issuedOK := parseInstant(value.IssuedAt)
	return value.Schema == CheckpointSchema &&
		(value.RegistryKind == RegistryDevelopment || value.RegistryKind == RegistrySaaS) &&
		value.RegistrySequence > 0 && value.ActiveScopeCount >= 0 &&
		validDigest(value.MerkleRootSHA256) && issuedOK && validKID(value.RegistryKID)
}
