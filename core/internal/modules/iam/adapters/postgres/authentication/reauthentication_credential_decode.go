package postgres

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (row *reauthenticationCredentialRow) decode(
	lookup authentication.ReauthenticationLookup,
) (authentication.ReauthenticationCredential, error) {
	defer row.clear()
	credential := &row.credential
	if len(row.challengeHMAC) != len(credential.ChallengeHMAC) ||
		len(row.contextSHA) != len(credential.ContextSHA256) ||
		len(row.purposeSHA) != len(credential.PurposeSHA256) ||
		len(row.originSHA) != len(credential.OriginSHA256) {
		return authentication.ReauthenticationCredential{}, errAuthenticationStorage
	}
	credential.ChallengeID = lookup.ChallengeID
	credential.UserID = lookup.UserID
	credential.SessionID = lookup.SessionID
	credential.Method = lookup.Method
	credential.AccessMode = authentication.AccessMode(row.accessMode)
	copy(credential.ChallengeHMAC[:], row.challengeHMAC)
	copy(credential.ContextSHA256[:], row.contextSHA)
	copy(credential.PurposeSHA256[:], row.purposeSHA)
	copy(credential.OriginSHA256[:], row.originSHA)
	if row.portalID != nil {
		credential.PortalBindingID = *row.portalID
	}
	if row.portalState != nil {
		credential.PortalBindingState = *row.portalState
	}
	if row.portalRevision != nil {
		credential.PortalBindingRevision = *row.portalRevision
	}
	var err error
	switch lookup.Method {
	case authentication.ReauthenticationPassword:
		err = row.decodePassword()
	case authentication.ReauthenticationWebAuthn:
		err = row.decodeWebAuthn()
	case authentication.ReauthenticationTOTP:
		err = row.decodeTOTP()
	default:
		err = errAuthenticationStorage
	}
	if err != nil {
		return authentication.ReauthenticationCredential{}, err
	}
	return *credential, nil
}
