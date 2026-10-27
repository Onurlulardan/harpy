package iam

import (
	"context"
	"crypto/subtle"

	cryptography "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/cryptography"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) issueChallenge(
	ctx context.Context, claims accesstoken.Claims,
	authenticatorID, challengeID uuid.UUID, kind string,
) (lifecycle.ChallengeEvidence, string, error) {
	if ctx == nil || ctx.Err() != nil {
		return lifecycle.ChallengeEvidence{}, "", ErrLifecycleHTTPInvalid
	}
	keyID, key, err := application.oneTime.Active()
	if err != nil {
		return lifecycle.ChallengeEvidence{}, "", ErrLifecycleHTTPUnavailable
	}
	defer clear(key[:])
	issued, err := cryptography.NewChallengeGenerator().Generate()
	if err != nil {
		return lifecycle.ChallengeEvidence{}, "", ErrLifecycleHTTPUnavailable
	}
	defer issued.Destroy()
	digest, err := issued.HMAC(key[:])
	if err != nil {
		return lifecycle.ChallengeEvidence{}, "", ErrLifecycleHTTPUnavailable
	}
	contextSHA := lifecycleChallengeContext(claims.Tenant(), claims.Subject(),
		claims.Session(), authenticatorID, challengeID, kind, application.rpID, application.origin)
	var wire string
	err = issued.RevealOnce(func(value []byte) error { wire = string(value); return nil })
	if err != nil {
		return lifecycle.ChallengeEvidence{}, "", ErrLifecycleHTTPUnavailable
	}
	return lifecycle.ChallengeEvidence{
		ID: challengeID, HMACKeyID: keyID, HMAC: digest, ContextSHA256: contextSHA,
	}, wire, nil
}

func (application *lifecycleHTTPApplication) presentedChallenge(
	claims accesstoken.Claims, authenticatorID, challengeID uuid.UUID,
	kind, keyID string, wire []byte,
) (lifecycle.ChallengeEvidence, string, error) {
	presented, err := cryptography.NewPresentedChallenge(wire)
	if err != nil {
		return lifecycle.ChallengeEvidence{}, "", ErrLifecycleHTTPInvalidCredential
	}
	defer presented.Destroy()
	key, err := application.oneTime.HMACKey(keyID)
	if err != nil {
		return lifecycle.ChallengeEvidence{}, "", ErrLifecycleHTTPUnavailable
	}
	defer clear(key[:])
	digest, err := presented.HMAC(key[:])
	if err != nil || subtle.ConstantTimeCompare(digest[:], make([]byte, 32)) == 1 {
		return lifecycle.ChallengeEvidence{}, "", ErrLifecycleHTTPInvalidCredential
	}
	value, err := presented.Wire()
	if err != nil {
		return lifecycle.ChallengeEvidence{}, "", ErrLifecycleHTTPInvalidCredential
	}
	defer clear(value)
	contextSHA := lifecycleChallengeContext(claims.Tenant(), claims.Subject(),
		claims.Session(), authenticatorID, challengeID, kind, application.rpID, application.origin)
	return lifecycle.ChallengeEvidence{
		ID: challengeID, HMACKeyID: keyID, HMAC: digest, ContextSHA256: contextSHA,
	}, string(value), nil
}
