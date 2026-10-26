package aicredentialv1

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

var labelPattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{0,63}$`)

func (aad AAD) validate() error {
	if _, err := uuid.Parse(aad.TenantID); err != nil {
		return ErrInvalidEnvelope
	}
	if _, err := uuid.Parse(aad.ConfigID); err != nil {
		return ErrInvalidEnvelope
	}
	if aad.ConfigRevision < 1 || !labelPattern.MatchString(aad.Provider) ||
		aad.Audience != Audience || !labelPattern.MatchString(aad.KeyID) {
		return ErrInvalidEnvelope
	}
	return nil
}

func aadBytes(aad AAD) ([]byte, error) {
	if err := aad.validate(); err != nil {
		return nil, err
	}
	document, err := json.Marshal(aad)
	if err != nil {
		return nil, ErrInvalidEnvelope
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 2048)
	if err != nil {
		return nil, ErrInvalidEnvelope
	}
	return canonical, nil
}

func Parse(document []byte) (Envelope, error) {
	if len(document) == 0 || len(document) > MaximumEnvelopeBytes {
		return Envelope{}, ErrInvalidEnvelope
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumEnvelopeBytes)
	if err != nil || !bytes.Equal(canonical, document) {
		return Envelope{}, ErrInvalidEnvelope
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var envelope Envelope
	if err := decoder.Decode(&envelope); err != nil || envelope.ContractVersion != ContractVersion ||
		envelope.AAD.validate() != nil {
		return Envelope{}, ErrInvalidEnvelope
	}
	if _, err := decodeExact(envelope.EphemeralPublicKey, 32); err != nil {
		return Envelope{}, err
	}
	if _, err := decodeExact(envelope.Nonce, 12); err != nil {
		return Envelope{}, err
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(envelope.Ciphertext)
	if err != nil || len(ciphertext) < 17 || len(ciphertext) > MaximumPlaintextBytes+16 {
		return Envelope{}, ErrInvalidEnvelope
	}
	return envelope, nil
}

func decodeExact(value string, size int) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) != size || base64.RawURLEncoding.EncodeToString(decoded) != value {
		clear(decoded)
		return nil, fmt.Errorf("%w", ErrInvalidEnvelope)
	}
	return decoded, nil
}
