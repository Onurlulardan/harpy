package artifact

import (
	"crypto/ed25519"
)

func SignBindingProof(
	value BindingProof,
	privateKey ed25519.PrivateKey,
) (VerifiedBindingProof, error) {
	if value.Signature != "" || !validBinding(value) ||
		len(privateKey) != ed25519.PrivateKeySize {
		return VerifiedBindingProof{}, ErrInvalidDocument
	}
	placeholder, err := canonicalValue(value)
	if err != nil {
		return VerifiedBindingProof{}, err
	}
	unsigned, err := unsignedArtifact(placeholder)
	if err != nil {
		return VerifiedBindingProof{}, err
	}
	value.Signature = encodeSignature(ed25519.Sign(
		privateKey, signatureInput(bindingDomain, unsigned),
	))
	bytes, err := canonicalValue(value)
	if err != nil {
		return VerifiedBindingProof{}, err
	}
	return VerifiedBindingProof{Value: value, Bytes: bytes, SHA256: digestHex(bytes)}, nil
}

func VerifyBindingProof(
	document []byte,
	publicKey ed25519.PublicKey,
) (VerifiedBindingProof, error) {
	var value BindingProof
	canonical, err := strictCanonicalDecode(document, &value)
	if err != nil || !validBinding(value) || len(publicKey) != ed25519.PublicKeySize {
		return VerifiedBindingProof{}, ErrInvalidDocument
	}
	signature, ok := decodeBase64URL(value.Signature, ed25519.SignatureSize)
	unsigned, unsignedErr := unsignedArtifact(canonical)
	if !ok || unsignedErr != nil ||
		!ed25519.Verify(publicKey, signatureInput(bindingDomain, unsigned), signature) {
		return VerifiedBindingProof{}, ErrInvalidSignature
	}
	return VerifiedBindingProof{
		Value: value, Bytes: canonical, SHA256: digestHex(canonical),
	}, nil
}
