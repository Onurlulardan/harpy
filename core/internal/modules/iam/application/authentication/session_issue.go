package authentication

import (
	"time"

	"github.com/google/uuid"
)

type SessionIssue struct {
	UserID                        uuid.UUID
	ExpectedUserRevision          int64
	ExpectedTokenVersion          int64
	ExpectedAuthorizationVersion  int64
	AccessModeID                  uuid.UUID
	AccessMode                    AccessMode
	ExpectedAccessModeRevision    int64
	PortalBindingID               uuid.UUID
	ExpectedPortalBindingRevision int64
	SessionID                     uuid.UUID
	FamilyID                      uuid.UUID
	TokenID                       uuid.UUID
	Device                        DeviceEvidence
	Selector                      [16]byte
	HashKeyID                     string
	SecretHMAC                    [32]byte
}

type SessionIssued struct {
	SessionRevision            int64
	FamilyRevision             int64
	TokenRevision              int64
	IssuedAt                   time.Time
	IdleExpiresAt              time.Time
	AbsoluteExpiresAt          time.Time
	RetainUntil                time.Time
	PlatformAuthorizationClock int64
	SecurityEpoch              [32]byte
}

func (issued SessionIssued) validate() error {
	if issued.SessionRevision != 1 || issued.FamilyRevision != 1 ||
		issued.TokenRevision != 1 || issued.IssuedAt.IsZero() ||
		issued.IdleExpiresAt.IsZero() || issued.AbsoluteExpiresAt.IsZero() ||
		issued.RetainUntil.IsZero() ||
		issued.PlatformAuthorizationClock < 0 ||
		issued.SecurityEpoch == [32]byte{} ||
		!issued.IdleExpiresAt.Equal(issued.IssuedAt.Add(30*24*time.Hour)) ||
		!issued.AbsoluteExpiresAt.Equal(issued.IssuedAt.Add(90*24*time.Hour)) ||
		!issued.RetainUntil.Equal(issued.AbsoluteExpiresAt.Add(24*time.Hour)) {
		return ErrAuthenticationStorage
	}
	return nil
}
