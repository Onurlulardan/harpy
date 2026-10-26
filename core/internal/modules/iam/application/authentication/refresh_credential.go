package authentication

import (
	"crypto/subtle"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type RefreshCredential struct {
	TokenID, FamilyID, SessionID, UserID  uuid.UUID
	TokenRevision, FamilyRevision         int64
	SessionRevision, UserRevision         int64
	TokenState, FamilyState, SessionState string
	UserStatus, IdentityKind              string
	TokenVersion, AuthorizationVersion    int64
	PlatformAuthorizationClock            int64
	SecurityEpoch                         [32]byte
	AccessModeID                          uuid.UUID
	AccessMode                            AccessMode
	AccessModeState                       string
	AccessModeRevision                    int64
	PortalBindingID                       uuid.UUID
	PortalBindingState                    string
	PortalBindingRevision                 int64
	Selector                              [16]byte
	HashKeyID                             string
	SecretHMAC                            [32]byte
	LocatorSourceSHA256                   [32]byte
	Sequence                              int64
	AuthTime                              time.Time
	ACR                                   string
	AMRPassword, AMRWebAuthn, AMRTOTP     bool
	AMRRecovery                           bool
	StepUpPurposeSHA256                   [32]byte
	HasStepUpPurpose                      bool
	IdleExpiresAt, AbsoluteExpiresAt      time.Time
	RetainUntil                           time.Time
}

func (credential RefreshCredential) validate(
	scope platformtx.CredentialTenantScope,
) error {
	binding := scope.LocatorBindingSHA256()
	if credential.TokenID == uuid.Nil || credential.FamilyID == uuid.Nil ||
		credential.SessionID == uuid.Nil || credential.UserID == uuid.Nil ||
		credential.TokenID != scope.SourceTokenID() ||
		credential.FamilyID != scope.SourceFamilyID() ||
		credential.Selector != scope.SourceSelector() ||
		credential.TokenRevision != scope.SourceRevision() ||
		subtle.ConstantTimeCompare(credential.LocatorSourceSHA256[:], binding[:]) != 1 ||
		credential.revisionsInvalid() || credential.statesInvalid() ||
		credential.contextInvalid() || credential.expiryInvalid() {
		return ErrAuthenticationStorage
	}
	return nil
}

func (credential RefreshCredential) revisionsInvalid() bool {
	return credential.TokenRevision <= 0 || credential.FamilyRevision <= 0 ||
		credential.SessionRevision <= 0 || credential.UserRevision <= 0 ||
		credential.TokenVersion <= 0 || credential.AuthorizationVersion <= 0 ||
		credential.PlatformAuthorizationClock < 0 ||
		credential.SecurityEpoch == [32]byte{} ||
		credential.AccessModeID == uuid.Nil || credential.AccessModeRevision <= 0 ||
		credential.Sequence <= 0 || !validHMACKeyID(credential.HashKeyID)
}

func (credential RefreshCredential) statesInvalid() bool {
	return !oneOf(credential.TokenState, "active", "consumed", "revoked", "expired") ||
		!oneOf(credential.FamilyState, "active", "revoked", "compromised", "expired") ||
		!oneOf(credential.SessionState, "active", "revoked", "expired") ||
		!oneOf(credential.UserStatus, "pending", "active", "suspended", "disabled", "tombstoned") ||
		credential.IdentityKind != "human" ||
		!oneOf(credential.AccessModeState, "pending", "active", "suspended", "retired") ||
		credential.bindingShapeInvalid()
}

func (credential RefreshCredential) bindingShapeInvalid() bool {
	if credential.AccessMode == AccessModeOperator {
		return credential.PortalBindingID != uuid.Nil ||
			credential.PortalBindingState != "" || credential.PortalBindingRevision != 0
	}
	return credential.AccessMode != AccessModePortal ||
		credential.PortalBindingID == uuid.Nil ||
		!oneOf(credential.PortalBindingState, "pending", "active", "suspended", "retired") ||
		credential.PortalBindingRevision <= 0
}

func (credential RefreshCredential) contextInvalid() bool {
	_, err := credential.accessContext()
	return err != nil
}

func (credential RefreshCredential) expiryInvalid() bool {
	return credential.AuthTime.IsZero() || credential.IdleExpiresAt.IsZero() ||
		credential.AbsoluteExpiresAt.IsZero() || credential.RetainUntil.IsZero() ||
		credential.AuthTime.After(credential.AbsoluteExpiresAt) ||
		credential.IdleExpiresAt.After(credential.AbsoluteExpiresAt) ||
		credential.RetainUntil.Before(credential.AbsoluteExpiresAt.Add(24*time.Hour))
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
