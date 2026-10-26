package iam

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/google/uuid"
)

type authHTTPReauthenticationChallenge struct {
	inner *authentication.ReauthenticationChallenge
}

func (challenge *authHTTPReauthenticationChallenge) ID() uuid.UUID {
	if challenge == nil || challenge.inner == nil {
		return uuid.Nil
	}
	return challenge.inner.ID
}

func (challenge *authHTTPReauthenticationChallenge) Method() AuthHTTPReauthenticationMethod {
	if challenge == nil || challenge.inner == nil {
		return ""
	}
	return AuthHTTPReauthenticationMethod(challenge.inner.Method)
}

func (challenge *authHTTPReauthenticationChallenge) ExpiresAt() time.Time {
	if challenge == nil || challenge.inner == nil {
		return time.Time{}
	}
	return challenge.inner.ExpiresAt
}

func (challenge *authHTTPReauthenticationChallenge) CredentialID() []byte {
	if challenge == nil || challenge.inner == nil {
		return nil
	}
	return append([]byte(nil), challenge.inner.CredentialID...)
}

func (challenge *authHTTPReauthenticationChallenge) RPID() string {
	if challenge == nil || challenge.inner == nil {
		return ""
	}
	return challenge.inner.RPID
}

func (challenge *authHTTPReauthenticationChallenge) UserVerificationRequired() bool {
	return challenge != nil && challenge.inner != nil &&
		challenge.inner.UserVerificationRequired
}

func (challenge *authHTTPReauthenticationChallenge) RevealOnce(
	consume func([]byte) error,
) error {
	if challenge == nil || challenge.inner == nil {
		return authentication.ErrAuthenticationContract
	}
	return challenge.inner.RevealOnce(consume)
}

func (challenge *authHTTPReauthenticationChallenge) Destroy() {
	if challenge == nil || challenge.inner == nil {
		return
	}
	challenge.inner.Destroy()
	challenge.inner = nil
}
