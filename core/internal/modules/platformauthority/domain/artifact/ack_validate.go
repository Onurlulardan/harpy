package artifact

func validAck(value Ack) bool {
	_, instantOK := parseInstant(value.IssuedAt)
	return value.Schema == AckSchema &&
		(value.RegistryKind == RegistryDevelopment || value.RegistryKind == RegistrySaaS) &&
		validUUID(value.RegistryBindingID) && validUUID(value.ScopeID) &&
		(value.ScopeKind == ScopeEnvironment || value.ScopeKind == ScopeTenant) &&
		value.RegistrySequence > 0 && validUUID(value.OperationID) &&
		validDigest(value.ProposalSHA256) && value.Generation > 0 &&
		validOptionalDigest(value.PreviousHeadSHA256) && validDigest(value.StateSHA256) &&
		instantOK && validKID(value.RegistryKID) &&
		(value.Generation == 1) == (value.PreviousHeadSHA256 == nil)
}

func ackMatchesProposal(value Ack, proposal SignedProposal) bool {
	return value.RegistryBindingID == proposal.Value.RegistryBindingID &&
		value.ScopeKind == proposal.Value.ScopeKind && value.ScopeID == proposal.Value.ScopeID &&
		value.OperationID == proposal.Value.OperationID &&
		value.ProposalSHA256 == proposal.SHA256 && value.Generation == proposal.Value.Generation &&
		optionalEqual(value.PreviousHeadSHA256, proposal.Value.PreviousHeadSHA256) &&
		value.StateSHA256 == proposal.Value.StateSHA256
}

func optionalEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
