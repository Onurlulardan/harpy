package authentication

import (
	"time"

	"github.com/google/uuid"
)

func (credential ReauthenticationCredential) stepUpClaims(
	tenantID uuid.UUID,
	client ClientID,
	jti uuid.UUID,
	issuedAt time.Time,
) AccessClaims {
	input := AccessClaims{
		Subject: credential.UserID, Session: credential.SessionID,
		Tenant: tenantID, Client: client, IssuedAt: issuedAt, JTI: jti,
		TokenVersion:               credential.TokenVersion,
		AuthorizationVersion:       credential.AuthorizationVersion,
		PlatformAuthorizationClock: credential.PlatformAuthorizationClock,
		SecurityEpoch:              credential.SecurityEpoch,
		AccessMode:                 credential.AccessMode, AuthenticationTime: issuedAt,
		Assurance:         AssuranceStepUp,
		Methods:           []AuthMethod{credential.Method.accessMethod()},
		StepUpPurposeHash: append([]byte(nil), credential.PurposeSHA256[:]...),
	}
	if credential.PortalBindingID != uuid.Nil {
		binding := credential.PortalBindingID
		input.PortalBinding = &binding
	}
	return input
}
