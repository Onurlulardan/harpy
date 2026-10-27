package api

import (
	"context"
	"errors"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	notificationsmodule "github.com/Onurlulardan/harpy/core/internal/modules/notifications"
)

type notificationsIAMAccess interface {
	Verify(context.Context, iammodule.NotificationsAccessRequest) (
		iammodule.NotificationsAccessDecision, error)
}

type notificationsAccessVerifier struct {
	iam       notificationsIAMAccess
	consumers notificationConsumerVerifier
}

func newNotificationsAccessVerifier(iam notificationsIAMAccess,
	consumers notificationConsumerVerifier,
) (*notificationsAccessVerifier, error) {
	if iam == nil || consumers.now == nil || consumers.keyring.Validate() != nil {
		return nil, fmt.Errorf("Notifications IAM access is required")
	}
	return &notificationsAccessVerifier{iam: iam, consumers: consumers}, nil
}

func (verifier *notificationsAccessVerifier) Verify(ctx context.Context,
	request notificationsmodule.HTTPAccessRequest,
) (notificationsmodule.HTTPAccess, error) {
	if verifier == nil || verifier.iam == nil {
		return notificationsmodule.HTTPAccess{}, notificationsmodule.ErrHTTPUnavailable
	}
	decision, err := verifier.iam.Verify(ctx, iammodule.NotificationsAccessRequest{
		AuthorizationValues: request.AuthorizationValues, CookieValues: request.CookieValues,
		TenantIDValues: request.TenantIDValues, TenantSlugValues: request.TenantSlugValues,
		CSRFValues: request.CSRFValues, OriginValues: request.OriginValues,
		Permission: string(request.Permission), Operation: request.Operation,
		RequireSession: request.RequireSession, RequireCSRF: request.RequireCSRF,
		RequireHTTPSOrigin: request.RequireHTTPSOrigin,
		RequireSecure:      request.RequireSecureTransport,
		RequireConsumer:    request.RequireSignedConsumer,
	})
	if err != nil {
		switch {
		case errors.Is(err, iammodule.ErrNotificationsAccessAuthentication):
			return notificationsmodule.HTTPAccess{}, notificationsmodule.ErrHTTPAuthentication
		case errors.Is(err, iammodule.ErrNotificationsAccessForbidden):
			return notificationsmodule.HTTPAccess{}, notificationsmodule.ErrHTTPForbidden
		default:
			return notificationsmodule.HTTPAccess{}, notificationsmodule.ErrHTTPUnavailable
		}
	}
	if request.RequireSignedConsumer {
		digest, verifyErr := verifier.consumers.Verify(
			request.ConsumerManifestValues, request.ConsumerSignatureValues)
		if verifyErr != nil {
			return notificationsmodule.HTTPAccess{}, notificationsmodule.ErrHTTPForbidden
		}
		return notificationsmodule.NewLegacyHTTPAccess(decision.PrincipalID(),
			decision.TenantID(), decision.TokenVersion(), decision.AuthorizationVersion(),
			decision.Origin(), digest, decision.SessionAuthenticated())
	}
	access, err := notificationsmodule.NewHTTPAccess(decision.PrincipalID(),
		decision.TenantID(), decision.TokenVersion(), decision.AuthorizationVersion(),
		request.Permission, request.Operation, decision.Origin(),
		decision.SessionAuthenticated(), decision.CSRFVerified())
	if err != nil {
		return notificationsmodule.HTTPAccess{}, notificationsmodule.ErrHTTPUnavailable
	}
	return access, nil
}

var _ notificationsmodule.HTTPAccessVerifier = (*notificationsAccessVerifier)(nil)
