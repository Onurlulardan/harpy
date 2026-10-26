package authentication

import (
	"context"
	"time"
)

type ChallengeGenerator interface {
	Generate() (IssuedChallenge, error)
}

type ChallengeKeyring interface {
	Active() (string, [32]byte, error)
	HMACKey(string) ([32]byte, error)
}

type RecoveryKeyring interface {
	HMACKey(string) ([32]byte, error)
}

type WebAuthnVerifier interface {
	Verify(WebAuthnAssertion, WebAuthnExpectation) (WebAuthnResult, error)
}

type TOTPVerifier interface {
	Verify(context.Context, TOTPCode, TOTPEnvelope, time.Time) (int64, error)
}

type SecurityCryptography struct {
	Challenges    ChallengeGenerator
	ChallengeKeys ChallengeKeyring
	RecoveryKeys  RecoveryKeyring
	WebAuthn      WebAuthnVerifier
	TOTP          TOTPVerifier
}

func (security SecurityCryptography) validate() error {
	if nilInterface(security.Challenges) || nilInterface(security.ChallengeKeys) ||
		nilInterface(security.RecoveryKeys) || nilInterface(security.WebAuthn) ||
		nilInterface(security.TOTP) {
		return ErrAuthenticationContract
	}
	return nil
}
