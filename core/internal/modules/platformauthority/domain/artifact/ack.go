package artifact

import (
	"crypto/ed25519"
	"time"
)

func SignAck(
	value Ack,
	proposal SignedProposal,
	privateKey ed25519.PrivateKey,
) (VerifiedAck, error) {
	if len(privateKey) != ed25519.PrivateKeySize || value.Signature != "" ||
		!validAck(value) || !ackMatchesProposal(value, proposal) {
		return VerifiedAck{}, ErrInvalidDocument
	}
	placeholder, err := canonicalValue(value)
	if err != nil {
		return VerifiedAck{}, err
	}
	unsigned, err := unsignedArtifact(placeholder)
	if err != nil {
		return VerifiedAck{}, err
	}
	value.Signature = encodeSignature(ed25519.Sign(
		privateKey, signatureInput(ackDomain, unsigned),
	))
	bytes, err := canonicalValue(value)
	if err != nil {
		return VerifiedAck{}, err
	}
	return VerifiedAck{Value: value, Bytes: bytes, SHA256: digestHex(bytes)}, nil
}

func VerifyAck(
	document []byte,
	proposal SignedProposal,
	trust VerifiedTrustBundle,
	now time.Time,
) (VerifiedAck, error) {
	ack, err := verifyAckDocument(document)
	if err != nil || !ackMatchesProposal(ack.Value, proposal) {
		return VerifiedAck{}, ErrInvalidDocument
	}
	if ack.Value.RegistryKind != trust.Kind {
		return VerifiedAck{}, ErrWrongRegistryKind
	}
	issuedAt, _ := parseInstant(ack.Value.IssuedAt)
	key, err := trust.resolve(ack.Value.RegistryKID, issuedAt, now, false)
	if err != nil {
		return VerifiedAck{}, err
	}
	if err := verifyAckWithKey(ack, key); err != nil {
		return VerifiedAck{}, err
	}
	return ack, nil
}

func VerifyAckSignature(
	document []byte,
	expectedKind RegistryKind,
	publicKey ed25519.PublicKey,
) (VerifiedAck, error) {
	ack, err := verifyAckDocument(document)
	if err != nil {
		return VerifiedAck{}, err
	}
	if ack.Value.RegistryKind != expectedKind {
		return VerifiedAck{}, ErrWrongRegistryKind
	}
	if err := verifyAckWithKey(ack, publicKey); err != nil {
		return VerifiedAck{}, err
	}
	return ack, nil
}

// VerifyAckPinned verifies an ack against an already authenticated exact trust
// bundle when the originating proposal is intentionally unavailable.
func VerifyAckPinned(
	document []byte,
	trust VerifiedTrustBundle,
	now time.Time,
) (VerifiedAck, error) {
	ack, err := verifyAckDocument(document)
	if err != nil {
		return VerifiedAck{}, err
	}
	if ack.Value.RegistryKind != trust.Kind {
		return VerifiedAck{}, ErrWrongRegistryKind
	}
	issuedAt, _ := parseInstant(ack.Value.IssuedAt)
	key, err := trust.resolve(ack.Value.RegistryKID, issuedAt, now, false)
	if err != nil {
		return VerifiedAck{}, err
	}
	if err = verifyAckWithKey(ack, key); err != nil {
		return VerifiedAck{}, err
	}
	return ack, nil
}

func verifyAckDocument(document []byte) (VerifiedAck, error) {
	var value Ack
	canonical, err := strictCanonicalDecode(document, &value)
	if err != nil || !validAck(value) {
		return VerifiedAck{}, ErrInvalidDocument
	}
	return VerifiedAck{Value: value, Bytes: canonical, SHA256: digestHex(canonical)}, nil
}

func ParseAck(document []byte) (VerifiedAck, error) {
	ack, err := verifyAckDocument(document)
	if err != nil {
		return VerifiedAck{}, err
	}
	if _, valid := decodeBase64URL(ack.Value.Signature, ed25519.SignatureSize); !valid {
		return VerifiedAck{}, ErrInvalidSignature
	}
	return ack, nil
}

func verifyAckWithKey(ack VerifiedAck, key ed25519.PublicKey) error {
	if len(key) != ed25519.PublicKeySize {
		return ErrUntrustedKey
	}
	signature, ok := decodeBase64URL(ack.Value.Signature, ed25519.SignatureSize)
	unsigned, unsignedErr := unsignedArtifact(ack.Bytes)
	if !ok || unsignedErr != nil ||
		!ed25519.Verify(key, signatureInput(ackDomain, unsigned), signature) {
		return ErrInvalidSignature
	}
	return nil
}
