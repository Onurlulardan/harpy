package billing

import (
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type VerifiedHTTPAccess = billingapi.Access

var (
	ErrHTTPAuthentication = billingapi.ErrAuthentication
	ErrHTTPForbidden      = billingapi.ErrForbidden
	ErrHTTPUnavailable    = billingapi.ErrUnavailable
)

func NewTenantVerifiedHTTPAccess(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission billingapi.Permission,
) (VerifiedHTTPAccess, error) {
	return billingapi.NewTenantAccessEvidence(
		principalID, tenantID, tokenVersion, authorizationVersion,
		authorizationClock, permission,
	)
}

func NewGlobalVerifiedHTTPAccess(
	principalID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission billingapi.Permission,
	capability platformtx.GlobalCommandCapability,
) (VerifiedHTTPAccess, error) {
	return billingapi.NewGlobalAccessEvidence(
		principalID, tokenVersion, authorizationVersion,
		authorizationClock, permission, capability,
	)
}

func NewGlobalCommandVerifiedHTTPAccess(
	principalID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission billingapi.Permission,
	capability platformtx.GlobalCommandAccess,
) (VerifiedHTTPAccess, error) {
	return billingapi.NewGlobalCommandAccessEvidence(
		principalID, tokenVersion, authorizationVersion,
		authorizationClock, permission, capability,
	)
}

func NewGlobalSnapshotVerifiedHTTPAccess(
	principalID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission billingapi.Permission,
	capability platformtx.GlobalSnapshotCapability,
) (VerifiedHTTPAccess, error) {
	return billingapi.NewGlobalSnapshotAccessEvidence(
		principalID, tokenVersion, authorizationVersion,
		authorizationClock, permission, capability,
	)
}
