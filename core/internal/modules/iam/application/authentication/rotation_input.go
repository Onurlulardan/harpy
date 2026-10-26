package authentication

import "github.com/google/uuid"

func (credential RefreshCredential) rotation(
	successorID uuid.UUID,
	successorSelector [16]byte,
	successorHashKeyID string,
	successorSecretHMAC [32]byte,
) RefreshRotation {
	return RefreshRotation{
		TokenID: credential.TokenID, FamilyID: credential.FamilyID,
		SessionID: credential.SessionID, UserID: credential.UserID,
		ExpectedTokenRevision:              credential.TokenRevision,
		ExpectedFamilyRevision:             credential.FamilyRevision,
		ExpectedSessionRevision:            credential.SessionRevision,
		ExpectedUserRevision:               credential.UserRevision,
		ExpectedTokenVersion:               credential.TokenVersion,
		ExpectedAuthorizationVersion:       credential.AuthorizationVersion,
		ExpectedPlatformAuthorizationClock: credential.PlatformAuthorizationClock,
		ExpectedSecurityEpoch:              credential.SecurityEpoch,
		AccessModeID:                       credential.AccessModeID,
		ExpectedAccessModeRevision:         credential.AccessModeRevision,
		PortalBindingID:                    credential.PortalBindingID,
		ExpectedPortalBindingRevision:      credential.PortalBindingRevision,
		SuccessorID:                        successorID, SuccessorSelector: successorSelector,
		SuccessorHashKeyID:  successorHashKeyID,
		SuccessorSecretHMAC: successorSecretHMAC,
	}
}
