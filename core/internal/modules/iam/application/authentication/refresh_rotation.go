package authentication

import "github.com/google/uuid"

type RefreshRotation struct {
	TokenID, FamilyID, SessionID, UserID uuid.UUID
	ExpectedTokenRevision                int64
	ExpectedFamilyRevision               int64
	ExpectedSessionRevision              int64
	ExpectedUserRevision                 int64
	ExpectedTokenVersion                 int64
	ExpectedAuthorizationVersion         int64
	ExpectedPlatformAuthorizationClock   int64
	ExpectedSecurityEpoch                [32]byte
	AccessModeID                         uuid.UUID
	ExpectedAccessModeRevision           int64
	PortalBindingID                      uuid.UUID
	ExpectedPortalBindingRevision        int64
	SuccessorID                          uuid.UUID
	SuccessorSelector                    [16]byte
	SuccessorHashKeyID                   string
	SuccessorSecretHMAC                  [32]byte
}
