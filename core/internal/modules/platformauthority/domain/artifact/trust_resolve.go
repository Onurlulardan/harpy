package artifact

import (
	"crypto/ed25519"
	"time"
)

func (bundle VerifiedTrustBundle) resolve(
	kid string,
	issuedAt time.Time,
	now time.Time,
	forNewSignature bool,
) (ed25519.PublicKey, error) {
	key, exists := bundle.keys[kid]
	if !exists || key.revoked || len(key.publicKey) != ed25519.PublicKeySize ||
		issuedAt.After(key.signUntil) || now.After(key.verifyUntil) ||
		(forNewSignature && !key.active) {
		return nil, ErrUntrustedKey
	}
	return key.publicKey, nil
}
