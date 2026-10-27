package artifact

import (
	"crypto/ed25519"
	"time"
)

func SignCheckpoint(
	value Checkpoint,
	privateKey ed25519.PrivateKey,
) (VerifiedCheckpoint, error) {
	if value.Signature != "" || !validCheckpoint(value) ||
		len(privateKey) != ed25519.PrivateKeySize {
		return VerifiedCheckpoint{}, ErrInvalidDocument
	}
	placeholder, err := canonicalValue(value)
	if err != nil {
		return VerifiedCheckpoint{}, err
	}
	unsigned, err := unsignedArtifact(placeholder)
	if err != nil {
		return VerifiedCheckpoint{}, err
	}
	value.Signature = encodeSignature(ed25519.Sign(
		privateKey, signatureInput(checkpointDomain, unsigned),
	))
	bytes, err := canonicalValue(value)
	if err != nil {
		return VerifiedCheckpoint{}, err
	}
	return VerifiedCheckpoint{Value: value, Bytes: bytes, SHA256: digestHex(bytes)}, nil
}

func VerifyCheckpoint(
	document []byte,
	trust VerifiedTrustBundle,
	now time.Time,
) (VerifiedCheckpoint, error) {
	var value Checkpoint
	canonical, err := strictCanonicalDecode(document, &value)
	if err != nil || !validCheckpoint(value) {
		return VerifiedCheckpoint{}, ErrInvalidDocument
	}
	if value.RegistryKind != trust.Kind {
		return VerifiedCheckpoint{}, ErrWrongRegistryKind
	}
	issued, _ := parseInstant(value.IssuedAt)
	key, err := trust.resolve(value.RegistryKID, issued, now, false)
	if err != nil {
		return VerifiedCheckpoint{}, err
	}
	signature, ok := decodeBase64URL(value.Signature, ed25519.SignatureSize)
	unsigned, unsignedErr := unsignedArtifact(canonical)
	if !ok || unsignedErr != nil ||
		!ed25519.Verify(key, signatureInput(checkpointDomain, unsigned), signature) {
		return VerifiedCheckpoint{}, ErrInvalidSignature
	}
	return VerifiedCheckpoint{
		Value: value, Bytes: canonical, SHA256: digestHex(canonical),
	}, nil
}

func VerifyCheckpointSignature(
	document []byte,
	expectedKind RegistryKind,
	publicKey ed25519.PublicKey,
) (VerifiedCheckpoint, error) {
	checkpoint, err := ParseCheckpoint(document)
	if err != nil {
		return VerifiedCheckpoint{}, err
	}
	if checkpoint.Value.RegistryKind != expectedKind {
		return VerifiedCheckpoint{}, ErrWrongRegistryKind
	}
	if len(publicKey) != ed25519.PublicKeySize {
		return VerifiedCheckpoint{}, ErrUntrustedKey
	}
	signature, ok := decodeBase64URL(checkpoint.Value.Signature, ed25519.SignatureSize)
	unsigned, unsignedErr := unsignedArtifact(checkpoint.Bytes)
	if !ok || unsignedErr != nil || !ed25519.Verify(publicKey,
		signatureInput(checkpointDomain, unsigned), signature) {
		return VerifiedCheckpoint{}, ErrInvalidSignature
	}
	return checkpoint, nil
}

func ParseCheckpoint(document []byte) (VerifiedCheckpoint, error) {
	var value Checkpoint
	canonical, err := strictCanonicalDecode(document, &value)
	if err != nil || !validCheckpoint(value) {
		return VerifiedCheckpoint{}, ErrInvalidDocument
	}
	if _, valid := decodeBase64URL(value.Signature, ed25519.SignatureSize); !valid {
		return VerifiedCheckpoint{}, ErrInvalidSignature
	}
	return VerifiedCheckpoint{Value: value, Bytes: canonical,
		SHA256: digestHex(canonical)}, nil
}
