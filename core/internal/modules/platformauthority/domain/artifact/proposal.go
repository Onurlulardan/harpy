package artifact

import (
	"crypto/ed25519"
	"time"
)

func SignProposal(
	value Proposal,
	state CanonicalState,
	policy ExpiryPolicy,
	privateKey ed25519.PrivateKey,
) (SignedProposal, error) {
	if state.Validate() != nil || len(privateKey) != ed25519.PrivateKeySize ||
		value.Signature != "" || value.StateSHA256 != state.SHA256 ||
		!proposalMatchesState(value, state.Value) || !validProposal(value, policy) {
		return SignedProposal{}, ErrInvalidDocument
	}
	prepared, err := PrepareProposal(value, state, policy)
	if err != nil {
		return SignedProposal{}, err
	}
	signature := ed25519.Sign(privateKey, prepared.SignatureInput)
	return AttachProposalSignature(prepared, signature)
}

func VerifyProposal(
	document []byte,
	state CanonicalState,
	policy ExpiryPolicy,
	now time.Time,
	publicKey ed25519.PublicKey,
) (SignedProposal, error) {
	proposal, err := VerifyProposalSignature(document, policy, now, publicKey)
	if err != nil {
		return SignedProposal{}, err
	}
	if state.Validate() != nil || proposal.Value.StateSHA256 != state.SHA256 ||
		!proposalMatchesState(proposal.Value, state.Value) {
		return SignedProposal{}, ErrInvalidDocument
	}
	return proposal, nil
}

func VerifyProposalSignature(
	document []byte,
	policy ExpiryPolicy,
	now time.Time,
	publicKey ed25519.PublicKey,
) (SignedProposal, error) {
	var value Proposal
	canonical, err := strictCanonicalDecode(document, &value)
	if err != nil || len(publicKey) != ed25519.PublicKeySize ||
		!validProposal(value, policy) {
		return SignedProposal{}, ErrInvalidDocument
	}
	issued, _ := parseInstant(value.IssuedAt)
	expires, _ := parseInstant(value.ExpiresAt)
	if now.Before(issued) || !now.Before(expires) {
		return SignedProposal{}, ErrExpired
	}
	signature, valid := decodeBase64URL(value.Signature, ed25519.SignatureSize)
	unsigned, unsignedErr := unsignedArtifact(canonical)
	if !valid || unsignedErr != nil ||
		!ed25519.Verify(publicKey, signatureInput(proposalDomain, unsigned), signature) {
		return SignedProposal{}, ErrInvalidSignature
	}
	return SignedProposal{Value: value, Bytes: canonical, SHA256: digestHex(canonical)}, nil
}

func ParseSignedProposal(document []byte) (SignedProposal, error) {
	var value Proposal
	canonical, err := strictCanonicalDecode(document, &value)
	if err != nil || (!validProposal(value, ExpiryOnline) &&
		!validProposal(value, ExpiryAirGap)) {
		return SignedProposal{}, ErrInvalidDocument
	}
	if _, valid := decodeBase64URL(value.Signature, ed25519.SignatureSize); !valid {
		return SignedProposal{}, ErrInvalidSignature
	}
	return SignedProposal{Value: value, Bytes: canonical, SHA256: digestHex(canonical)}, nil
}

func VerifyStoredProposalSignature(
	document []byte,
	publicKey ed25519.PublicKey,
) (SignedProposal, error) {
	parsed, err := ParseSignedProposal(document)
	if err != nil {
		return SignedProposal{}, err
	}
	issuedAt, err := time.Parse(time.RFC3339, parsed.Value.IssuedAt)
	if err != nil {
		return SignedProposal{}, ErrInvalidDocument
	}
	policy := ExpiryOnline
	if expiresAt, parseErr := time.Parse(time.RFC3339, parsed.Value.ExpiresAt); parseErr != nil {
		return SignedProposal{}, ErrInvalidDocument
	} else if expiresAt.Sub(issuedAt) == 7*24*time.Hour {
		policy = ExpiryAirGap
	}
	return VerifyProposalSignature(document, policy, issuedAt, publicKey)
}

func proposalMatchesState(value Proposal, state State) bool {
	common := value.EnvironmentID == state.EnvironmentID &&
		value.DeploymentMode == state.DeploymentMode &&
		value.RegistryBindingID == state.RegistryBindingID &&
		value.ScopeKind == state.ScopeKind && value.ScopeID == state.ScopeID &&
		optionalDigestEqual(value.RecoveryTrustBundleSHA256, state.RecoveryTrustBundleSHA256)
	if !common {
		return false
	}
	if value.Purpose != "proposal_signer_rotation" {
		return value.SignerKID == state.ProposalSignerKID
	}
	return value.SignerRotation != nil && value.SignerKID != state.ProposalSignerKID &&
		value.SignerRotation.NewSignerKID == state.ProposalSignerKID &&
		state.ProposalSignerGeneration == value.SignerRotation.PreviousSignerGeneration+1
}

func optionalDigestEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
