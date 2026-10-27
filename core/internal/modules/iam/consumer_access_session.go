package iam

import (
	"context"
	"errors"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/sessionview"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var (
	errConsumerSessionRejected    = errors.New("IAM consumer session was rejected")
	errConsumerSessionUnavailable = errors.New("IAM consumer session is unavailable")
)

func (runtime *BillingAccessRuntime) readConsumerSession(
	ctx context.Context, claims accesstoken.Claims,
) (sessionview.View, error) {
	if runtime == nil || runtime.sessions == nil {
		return sessionview.View{}, errConsumerSessionUnavailable
	}
	generations, err := runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, claims.Tenant(), currentSessionDomains,
	)
	if err != nil || !validIAMLayoutSnapshotGenerations(generations) {
		return sessionview.View{}, errConsumerSessionUnavailable
	}
	request, err := buildCurrentSessionScope(claims.Subject(), claims.Tenant(), generations)
	if err != nil {
		return sessionview.View{}, errConsumerSessionUnavailable
	}
	var view sessionview.View
	err = runtime.coordinator.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			if runtime.epoch.Assert(scoped, scope, claims.SecurityEpoch()) != nil {
				return errConsumerSessionRejected
			}
			var readErr error
			view, readErr = runtime.sessions.Read(
				scoped, scope, claims.Subject(), claims.Session(),
			)
			return readErr
		})
	if err != nil {
		if errors.Is(err, sessionview.ErrRejected) ||
			errors.Is(err, errConsumerSessionRejected) {
			return sessionview.View{}, errConsumerSessionRejected
		}
		return sessionview.View{}, errConsumerSessionUnavailable
	}
	if view.UserID != claims.Subject() || view.TenantID != claims.Tenant() ||
		view.SessionID != claims.Session() ||
		view.TokenVersion != claims.TokenVersion() ||
		view.AuthorizationVersion != claims.AuthorizationVersion() ||
		view.PlatformAuthorizationClock != claims.PlatformAuthorizationClock() ||
		view.SecurityEpoch != claims.SecurityEpoch() ||
		string(view.ClientID) != string(claims.Client()) ||
		string(view.AccessMode) != string(claims.AccessMode()) {
		return sessionview.View{}, errConsumerSessionRejected
	}
	return view, nil
}
