package authentication

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"
)

func (authenticator Authenticator) verifyReauthenticationProof(
	ctx context.Context,
	request ReauthenticateRequest,
	credential ReauthenticationCredential,
	now time.Time,
) (ReauthenticationProof, bool, error) {
	key, err := authenticator.security.ChallengeKeys.HMACKey(credential.HMACKeyID)
	if err != nil {
		return ReauthenticationProof{}, false, ErrAuthenticationCrypto
	}
	presented, err := request.Challenge.HMAC(key[:])
	clear(key[:])
	if err != nil {
		return ReauthenticationProof{}, false, ErrAuthenticationCrypto
	}
	defer clear(presented[:])
	if subtle.ConstantTimeCompare(presented[:], credential.ChallengeHMAC[:]) != 1 {
		return ReauthenticationProof{}, false, ErrInvalidCredentials
	}
	switch credential.Method {
	case ReauthenticationPassword:
		_, err = authenticator.verifyPassword(request.Password, credential.Password, now)
		return ReauthenticationProof{Method: credential.Method}, false, err
	case ReauthenticationWebAuthn:
		return authenticator.verifyWebAuthn(request, credential)
	case ReauthenticationTOTP:
		step, verifyErr := authenticator.security.TOTP.Verify(
			ctx, request.TOTP, credential.TOTP, now,
		)
		if verifyErr != nil {
			return ReauthenticationProof{}, false, verifyErr
		}
		if credential.TOTP.LastAcceptedTimeStep != nil &&
			step <= *credential.TOTP.LastAcceptedTimeStep {
			return ReauthenticationProof{}, false, ErrInvalidCredentials
		}
		return ReauthenticationProof{Method: credential.Method, TOTPTimeStep: step}, false, nil
	default:
		return ReauthenticationProof{}, false, ErrInvalidCredentials
	}
}

func proofFailure(errorValue error) (invalid bool) {
	return errors.Is(errorValue, ErrInvalidCredentials)
}
