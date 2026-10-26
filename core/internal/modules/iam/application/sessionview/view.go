package sessionview

import (
	"slices"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/sessioncapability"
	"github.com/google/uuid"
)

type ClientID string
type AccessMode string

const (
	ClientOperatorConsole ClientID   = "harpy-operator-console-v1"
	ClientPortal          ClientID   = "harpy-portal-v1"
	AccessModeOperator    AccessMode = "operator_console"
	AccessModePortal      AccessMode = "portal"
)

type View struct {
	UserID, TenantID, SessionID        uuid.UUID
	TenantSlug                         string
	Email, DisplayName                 string
	ClientID                           ClientID
	AccessMode                         AccessMode
	PortalBindingID                    uuid.UUID
	PlatformSubjectID                  uuid.UUID
	TokenVersion, AuthorizationVersion int64
	PlatformAuthorizationClock         int64
	PasswordCredentialRevision         *int64
	SecurityEpoch                      [32]byte
	DeploymentMode                     sessioncapability.DeploymentMode
	TenantPurpose                      sessioncapability.TenantPurpose
	PlatformCapabilities               []string
	TenantCapabilities                 []string
}

func (view View) valid(scopeTenantID uuid.UUID, userID, sessionID uuid.UUID) bool {
	if view.UserID != userID || view.TenantID != scopeTenantID ||
		view.SessionID != sessionID || view.TenantSlug == "" ||
		strings.TrimSpace(view.Email) == "" ||
		strings.TrimSpace(view.DisplayName) == "" ||
		view.TokenVersion <= 0 || view.AuthorizationVersion < 0 ||
		view.PlatformAuthorizationClock < 0 || view.SecurityEpoch == [32]byte{} ||
		(view.PasswordCredentialRevision != nil &&
			*view.PasswordCredentialRevision < 1) ||
		!view.TenantPurpose.ValidFor(view.DeploymentMode) ||
		!slices.IsSorted(view.TenantCapabilities) {
		return false
	}
	for index, capability := range view.TenantCapabilities {
		if !sessioncapability.ValidTenantCapability(capability) ||
			index > 0 && view.TenantCapabilities[index-1] == capability {
			return false
		}
	}
	if view.TenantPurpose == sessioncapability.PurposeSaaSCustomer {
		if view.PlatformSubjectID != uuid.Nil || view.PlatformAuthorizationClock != 0 {
			return false
		}
	} else if view.PlatformAuthorizationClock < 1 {
		return false
	}
	switch view.AccessMode {
	case AccessModeOperator:
		return view.ClientID == ClientOperatorConsole &&
			view.PortalBindingID == uuid.Nil
	case AccessModePortal:
		return view.ClientID == ClientPortal &&
			view.PortalBindingID != uuid.Nil
	default:
		return false
	}
}
