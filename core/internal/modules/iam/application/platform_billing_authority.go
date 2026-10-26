package application

import (
	"strings"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/sessioncapability"
	"github.com/google/uuid"
)

// PlatformBillingAuthority is sealed proof of an exact SaaS platform-IAM decision.
// Tenant RBAC decisions cannot construct it outside this package.
type PlatformBillingAuthority struct {
	actorID              uuid.UUID
	capability           string
	platformClock        int64
	authorizationVersion int64
	valid                bool
}

func NewPlatformBillingAuthority(
	decision platformdomain.Decision,
	actorID uuid.UUID,
	capability string,
	authorizationVersion int64,
) (PlatformBillingAuthority, error) {
	decisionActor := uuid.UUID(decision.SubjectID)
	valid := decision.Allowed && decision.Reason == "granted" &&
		decision.DeploymentMode == string(sessioncapability.ModeSaaS) &&
		decisionActor == actorID && actorID != uuid.Nil &&
		decision.Capability == capability &&
		strings.HasPrefix(capability, "platform.billing.") &&
		sessioncapability.ModeAllowsPlatform(sessioncapability.ModeSaaS, capability) &&
		decision.PlatformClock > 0 && authorizationVersion > 0 &&
		decision.AuthorizationVersion == authorizationVersion
	if !valid {
		return PlatformBillingAuthority{}, ErrGlobalBillingAuthority
	}
	return PlatformBillingAuthority{
		actorID: actorID, capability: capability,
		platformClock:        decision.PlatformClock,
		authorizationVersion: authorizationVersion, valid: true,
	}, nil
}

func (authority PlatformBillingAuthority) allows(command bool) bool {
	if !authority.valid || authority.actorID == uuid.Nil ||
		authority.platformClock < 1 || authority.authorizationVersion < 1 {
		return false
	}
	if command {
		return strings.HasSuffix(authority.capability, ".manage")
	}
	return strings.HasSuffix(authority.capability, ".read")
}
