package licenseauthority

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func Verify(
	ctx context.Context,
	document []byte,
	binding Binding,
	now time.Time,
	keys KeyResolver,
) (VerifiedSnapshot, error) {
	if len(document) == 0 || len(document) > MaximumDocumentBytes || keys == nil ||
		now.IsZero() || now.Location() != time.UTC || !validBinding(binding) {
		return VerifiedSnapshot{}, ErrInvalidDocument
	}
	canonicalEnvelope, err := canonicaljsonv1.CanonicalizeObject(
		document, MaximumDocumentBytes,
	)
	if err != nil || !bytes.Equal(document, canonicalEnvelope) {
		return VerifiedSnapshot{}, ErrInvalidDocument
	}
	var envelope wireEnvelope
	if strictDecode(canonicalEnvelope, &envelope) != nil ||
		!validEnvelope(envelope) {
		return VerifiedSnapshot{}, ErrInvalidDocument
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		envelope.Payload, MaximumDocumentBytes,
	)
	if err != nil {
		return VerifiedSnapshot{}, ErrInvalidDocument
	}
	if !bytes.Equal(envelope.Payload, canonical) {
		return VerifiedSnapshot{}, ErrInvalidDocument
	}
	digest := sha256.Sum256(canonical)
	if hex.EncodeToString(digest[:]) != envelope.PayloadSHA256 {
		return VerifiedSnapshot{}, ErrInvalidDocument
	}
	payload, err := parsePayload(canonical)
	if err != nil {
		return VerifiedSnapshot{}, err
	}
	key, err := keys.LicenseKey(
		ctx, payload.wire.Issuer, envelope.KeyID, payload.issuedAt, now,
	)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return VerifiedSnapshot{}, ErrUntrustedKey
	}
	signature, err := base64.RawURLEncoding.DecodeString(envelope.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize ||
		base64.RawURLEncoding.EncodeToString(signature) != envelope.Signature ||
		!ed25519.Verify(key, signatureMessage(canonical), signature) {
		return VerifiedSnapshot{}, ErrInvalidSignature
	}
	if !bindingMatches(binding, payload) {
		return VerifiedSnapshot{}, ErrWrongBinding
	}
	if now.Before(payload.notBefore) {
		return VerifiedSnapshot{}, ErrNotYetValid
	}
	return verifiedFrom(payload, envelope, document, canonical), nil
}

func validEnvelope(value wireEnvelope) bool {
	return value.Schema == EnvelopeSchema && value.Algorithm == SignatureAlgorithm &&
		value.Canonicalization == Canonicalization && keyIDPattern.MatchString(value.KeyID) &&
		digestPattern.MatchString(value.PayloadSHA256) && len(value.Payload) > 0 &&
		len(value.Signature) == 86
}

func validBinding(value Binding) bool {
	return value.LicenseID != [16]byte{} && value.LicenseeOrganizationID != [16]byte{} &&
		value.EnvironmentID != [16]byte{} && validDelivery(value.DeliveryKind)
}

func bindingMatches(binding Binding, payload parsedPayload) bool {
	return binding.LicenseID == payload.licenseID &&
		binding.LicenseeOrganizationID == payload.organizationID &&
		binding.EnvironmentID == payload.environmentID &&
		binding.DeliveryKind == payload.wire.DeliveryKind
}

func signatureMessage(canonical []byte) []byte {
	message := make([]byte, 0, len(signatureDomain)+len(canonical))
	message = append(message, signatureDomain...)
	return append(message, canonical...)
}
