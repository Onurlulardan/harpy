package iam

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type reauthenticationSession struct {
	userID, tenantID, sessionID, portalBindingID uuid.UUID
	tokenVersion, authorizationVersion           int64
	accessMode                                   authentication.AccessMode
	client                                       authentication.ClientID
}

func reauthenticationSessionFromClaims(
	claims accesstoken.Claims,
) (reauthenticationSession, error) {
	session := reauthenticationSession{userID: claims.Subject(), tenantID: claims.Tenant(),
		sessionID: claims.Session(), tokenVersion: claims.TokenVersion(),
		authorizationVersion: claims.AuthorizationVersion(),
		accessMode:           authentication.AccessMode(claims.AccessMode()),
		client:               authentication.ClientID(claims.Client())}
	if binding, present := claims.PortalBinding(); present {
		session.portalBindingID = binding
	}
	bindingSHA, err := authentication.ReauthenticationBindingSHA256(
		session.tenantID, session.userID, session.sessionID,
		session.tokenVersion, session.authorizationVersion,
		session.accessMode, session.portalBindingID,
	)
	if err != nil {
		return reauthenticationSession{}, err
	}
	if _, err := platformtx.NewSessionCredentialLookup(
		"iam.authentication.reauthentication", session.tenantID, bindingSHA,
	); err != nil {
		return reauthenticationSession{}, err
	}
	return session, nil
}

func (session reauthenticationSession) lookup() (platformtx.CredentialLookup, error) {
	binding, err := authentication.ReauthenticationBindingSHA256(
		session.tenantID, session.userID, session.sessionID,
		session.tokenVersion, session.authorizationVersion,
		session.accessMode, session.portalBindingID,
	)
	if err != nil {
		return platformtx.CredentialLookup{}, err
	}
	return platformtx.NewSessionCredentialLookup(
		"iam.authentication.reauthentication", session.tenantID, binding,
	)
}
