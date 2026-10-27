package fcartv1

import "io"

func (keyring *EnvelopeKeyring) wrap(
	identity KeyIdentity,
	key KEK,
	dek DEK,
) (KeyEnvelope, error) {
	aead, err := keyAEAD(key)
	if err != nil {
		return KeyEnvelope{}, err
	}
	var nonce [NonceBytes]byte
	if _, err := io.ReadFull(keyring.random, nonce[:]); err != nil {
		return KeyEnvelope{}, ErrRandomSource
	}
	if nonce == [NonceBytes]byte{} {
		return KeyEnvelope{}, ErrRandomSource
	}
	aad, err := identity.aad(key.id)
	if err != nil {
		return KeyEnvelope{}, err
	}
	envelope := KeyEnvelope{KEKID: key.id, Nonce: nonce,
		Ciphertext: aead.Seal(nil, nonce[:], dek[:], aad)}
	if err := envelope.Validate(); err != nil {
		return KeyEnvelope{}, err
	}
	return envelope, nil
}
