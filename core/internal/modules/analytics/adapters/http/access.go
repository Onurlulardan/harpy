package analyticshttp

import (
	"time"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	"github.com/google/uuid"
)

var (
	ErrAuthentication = httptransport.ErrAuthentication
	ErrForbidden      = httptransport.ErrForbidden
	ErrInvalid        = httptransport.ErrInvalid
	ErrNotFound       = httptransport.ErrNotFound
	ErrConflict       = httptransport.ErrConflict
	ErrRateLimited    = httptransport.ErrRateLimited
	ErrUnavailable    = httptransport.ErrUnavailable
)

type Permission = httptransport.Permission
type AccessRequest = httptransport.AccessRequest
type AccessVerifier = httptransport.AccessVerifier
type Access = httptransport.Access
type ErrorSink = httptransport.ErrorSink
type CorrelationResolver = httptransport.CorrelationResolver
type TrustedClientIPResolver = httptransport.TrustedClientIPResolver

func NewAccess(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	stepUpAuthenticatedAt time.Time, permission Permission, operation string,
) (Access, error) {
	return httptransport.NewAccess(principalID, tenantID, tokenVersion,
		authorizationVersion, authorizationClock, stepUpAuthenticatedAt, permission, operation)
}
