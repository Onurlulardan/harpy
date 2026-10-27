package cryptography

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

type AccessIssuer struct {
	profile accesstoken.Profile
	signer  *accesstoken.Signer
}

func NewAccessIssuer(
	profile accesstoken.Profile,
	signer *accesstoken.Signer,
) (AccessIssuer, error) {
	if profile.Issuer() == "" || profile.Audience() == "" || signer == nil {
		return AccessIssuer{}, authentication.ErrAuthenticationContract
	}
	return AccessIssuer{profile: profile, signer: signer}, nil
}

func (issuer AccessIssuer) Issue(input authentication.AccessClaims) (string, error) {
	methods := make([]accesstoken.AuthMethod, len(input.Methods))
	for index, method := range input.Methods {
		methods[index] = accesstoken.AuthMethod(method)
	}
	claims, err := accesstoken.NewClaims(issuer.profile, accesstoken.ClaimsInput{
		Subject: input.Subject, Session: input.Session, Tenant: input.Tenant,
		Client: accesstoken.ClientID(input.Client), IssuedAt: input.IssuedAt,
		JTI: input.JTI, TokenVersion: input.TokenVersion,
		AuthorizationVersion:       input.AuthorizationVersion,
		PlatformAuthorizationClock: input.PlatformAuthorizationClock,
		SecurityEpoch:              input.SecurityEpoch,
		AccessMode:                 accesstoken.AccessMode(input.AccessMode),
		PortalBinding:              input.PortalBinding, AuthenticationTime: input.AuthenticationTime,
		Assurance: accesstoken.Assurance(input.Assurance), Methods: methods,
		StepUpPurposeHash: append([]byte(nil), input.StepUpPurposeHash...),
	})
	if err != nil || issuer.signer == nil {
		return "", authentication.ErrAuthenticationCrypto
	}
	return issuer.signer.Sign(claims)
}
