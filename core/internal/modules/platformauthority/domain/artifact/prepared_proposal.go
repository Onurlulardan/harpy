package artifact

type PreparedProposal struct {
	Value          Proposal
	State          CanonicalState
	Policy         ExpiryPolicy
	UnsignedBytes  []byte
	SignatureInput []byte
}

func PrepareProposal(
	value Proposal,
	state CanonicalState,
	policy ExpiryPolicy,
) (PreparedProposal, error) {
	if value.Signature != "" || state.Validate() != nil ||
		value.StateSHA256 != state.SHA256 || !proposalMatchesState(value, state.Value) ||
		!validProposal(value, policy) {
		return PreparedProposal{}, ErrInvalidDocument
	}
	placeholder, err := canonicalValue(value)
	if err != nil {
		return PreparedProposal{}, err
	}
	unsigned, err := unsignedArtifact(placeholder)
	if err != nil {
		return PreparedProposal{}, err
	}
	return PreparedProposal{
		Value: value, State: state, Policy: policy, UnsignedBytes: unsigned,
		SignatureInput: signatureInput(proposalDomain, unsigned),
	}, nil
}

func AttachProposalSignature(
	prepared PreparedProposal,
	signature []byte,
) (SignedProposal, error) {
	if len(signature) != 64 {
		return SignedProposal{}, ErrInvalidSignature
	}
	value := prepared.Value
	value.Signature = encodeSignature(signature)
	bytes, err := canonicalValue(value)
	if err != nil {
		return SignedProposal{}, err
	}
	return SignedProposal{Value: value, Bytes: bytes, SHA256: digestHex(bytes)}, nil
}
