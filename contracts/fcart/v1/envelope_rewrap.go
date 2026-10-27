package fcartv1

import "context"

func (keyring *EnvelopeKeyring) CurrentKEKID() string {
	if keyring == nil {
		return ""
	}
	return keyring.current.id
}

func (keyring *EnvelopeKeyring) SupportsKEK(id string) bool {
	if keyring == nil || id == "" {
		return false
	}
	_, found := keyring.keys[id]
	return found
}

func (keyring *EnvelopeKeyring) Rewrap(
	ctx context.Context,
	identity KeyIdentity,
	envelope KeyEnvelope,
) (KeyEnvelope, bool, error) {
	if keyring == nil || ctx == nil || identity.Validate() != nil ||
		envelope.Validate() != nil {
		return KeyEnvelope{}, false, ErrInvalidKeyEnvelope
	}
	dek, err := keyring.Unwrap(ctx, identity, envelope)
	if err != nil {
		return KeyEnvelope{}, false, err
	}
	defer clear(dek[:])
	if envelope.KEKID == keyring.current.id {
		return envelope.OwnedCopy(), false, nil
	}
	rewrapped, err := keyring.wrap(identity, keyring.current, dek)
	if err != nil {
		return KeyEnvelope{}, false, err
	}
	return rewrapped, true, nil
}
