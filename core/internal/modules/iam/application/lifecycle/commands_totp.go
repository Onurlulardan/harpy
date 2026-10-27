package lifecycle

import "github.com/google/uuid"

type TOTPStartCommand struct {
	Evidence                             MutationEvidence
	AuthenticatorID, UserID, SessionID   uuid.UUID
	DisplayLabel, KMSKeyID               string
	Nonce, Ciphertext, AuthenticationTag []byte
	AADSHA256                            [32]byte
	Challenge                            ChallengeEvidence
}

type TOTPReadQuery struct {
	AuthenticatorID, ChallengeID uuid.UUID
	ActorID, PermissionID        uuid.UUID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
}

type TOTPCompleteCommand struct {
	Evidence                      MutationEvidence
	AuthenticatorID               uuid.UUID
	Challenge                     ChallengeEvidence
	ExpectedAuthenticatorRevision int64
	ExpectedChallengeRevision     int64
	AcceptedTimeStep              int64
}

type RecoveryCodeIssue struct {
	ID         uuid.UUID `json:"id"`
	HMACSHA256 string    `json:"hmac_sha256"`
}

type RecoveryIssueCommand struct {
	Evidence                MutationEvidence
	AuthenticatorID, UserID uuid.UUID
	DisplayLabel, HMACKeyID string
	Codes                   []RecoveryCodeIssue
}

type RecoveryListQuery struct {
	UserID                       uuid.UUID
	ActorID, PermissionID        uuid.UUID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
}
