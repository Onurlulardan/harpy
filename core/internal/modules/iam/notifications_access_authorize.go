package iam

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *NotificationsAccessRuntime) authorize(ctx context.Context,
	claims accesstoken.Claims, target notificationsAccessTarget, now time.Time,
) (authorization.Decision, [16]byte, error) {
	request, err := authorization.NewRequest(target.resource, target.action,
		claims.TokenVersion(), claims.AuthorizationVersion())
	if err != nil {
		return authorization.Decision{}, [16]byte{}, ErrNotificationsAccessForbidden
	}
	scopeRequest, err := iamHTTPSnapshotScopeRequest(ctx, runtime.shared.pool, claims)
	if err != nil {
		return authorization.Decision{}, [16]byte{}, ErrNotificationsAccessUnavailable
	}
	var result authorization.Decision
	var selector [16]byte
	err = runtime.shared.coordinator.InTenant(ctx, scopeRequest,
		func(scopeContext context.Context, scope platformtx.TenantScope) error {
			decision, authorizeErr := runtime.shared.authorizer.Authorize(
				scopeContext, scope, request)
			if authorizeErr != nil {
				return fmt.Errorf("%w: %w", ErrNotificationsAccessUnavailable, authorizeErr)
			}
			if validateErr := decision.Validate(); validateErr != nil {
				return fmt.Errorf("%w: %w", ErrNotificationsAccessUnavailable, validateErr)
			}
			if !decision.Allowed() {
				return ErrNotificationsAccessForbidden
			}
			if _, available := decision.PermissionID(); !available ||
				decision.AuthorizationClock() < 1 {
				return ErrNotificationsAccessUnavailable
			}
			if target.session {
				selector, authorizeErr = runtime.sessions.SessionSelector(scopeContext,
					scope, claims.Subject(), claims.Session(), now)
				if authorizeErr != nil {
					return ErrNotificationsAccessAuthentication
				}
			}
			result = decision
			return nil
		})
	if err != nil {
		if errors.Is(err, ErrNotificationsAccessForbidden) {
			return authorization.Decision{}, [16]byte{}, ErrNotificationsAccessForbidden
		}
		if errors.Is(err, ErrNotificationsAccessAuthentication) {
			return authorization.Decision{}, [16]byte{}, ErrNotificationsAccessAuthentication
		}
		if errors.Is(err, ErrNotificationsAccessUnavailable) {
			return authorization.Decision{}, [16]byte{}, err
		}
		return authorization.Decision{}, [16]byte{}, ErrNotificationsAccessUnavailable
	}
	return result, selector, nil
}
