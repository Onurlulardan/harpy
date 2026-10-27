package lifecycle

import (
	"time"

	"github.com/google/uuid"
)

type Authenticator struct {
	ID, UserID                        uuid.UUID
	Kind, DisplayLabel, State         string
	Revision                          int64
	CreatedAt, UpdatedAt, ActivatedAt time.Time
	LastUsedAt                        time.Time
}

type RecoveryCode struct {
	AuthenticatorID                  uuid.UUID
	AuthenticatorRevision            int64
	CodeID                           uuid.UUID
	Position                         int16
	State                            string
	CreatedAt, ConsumedAt, RevokedAt time.Time
}

type Session struct {
	ID, PortalBindingID                   uuid.UUID
	AccessMode, DeviceKind                string
	BrowserFamily, OSFamily               string
	Revision                              int64
	CreatedAt, LastSeenAt, AbsoluteExpiry time.Time
}

type PasswordCredential struct {
	Algorithm             string
	Salt, Tag             []byte
	MemoryKiB, Iterations int
	Parallelism           int16
	LegacyBcryptHash      string
	Version, Revision     int64
	CreatedAt             time.Time
}

type TOTPEnrollment struct {
	UserID                               uuid.UUID
	KMSKeyID                             string
	Nonce, Ciphertext, AuthenticationTag []byte
	AADSHA256                            [32]byte
	AuthenticatorRevision                int64
	ChallengeHMACKeyID                   string
	ChallengeHMAC, ContextSHA256         [32]byte
	ChallengeRevision                    int64
	ExpiresAt                            time.Time
}
