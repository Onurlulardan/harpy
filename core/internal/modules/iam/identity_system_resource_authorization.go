package iam

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type identitySystemResourceAuthorization struct {
	application *identityHTTPApplication
	ctx         context.Context
	scope       platformtx.TenantScope
	evidence    identityPublicEvidence
	resourceKey string
	clock       int64
	clockSet    bool
}

func newIdentitySystemResourceAuthorization(
	application *identityHTTPApplication,
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityPublicEvidence,
	resource publicresource.Resource,
) (*identitySystemResourceAuthorization, error) {
	key, err := resource.AuthorityResource()
	if application == nil || err != nil {
		return nil, publicresource.ErrContract
	}
	return &identitySystemResourceAuthorization{
		application: application, ctx: ctx, scope: scope,
		evidence: evidence, resourceKey: key,
	}, nil
}

func (session *identitySystemResourceAuthorization) decision(
	action string,
) (authorization.Decision, error) {
	request, err := authorization.NewRequest(
		session.resourceKey, action,
		session.evidence.TokenVersion(), session.evidence.AuthorizationVersion(),
	)
	if err != nil {
		return authorization.Decision{}, publicresource.ErrContract
	}
	decision, err := session.application.authorizer.Authorize(
		session.ctx, session.scope, request,
	)
	if err != nil || decision.Validate() != nil ||
		decision.AuthorizationVersion() != session.evidence.AuthorizationVersion() ||
		session.clockSet && decision.AuthorizationClock() != session.clock {
		return authorization.Decision{}, publicresource.ErrUnavailable
	}
	session.clock, session.clockSet = decision.AuthorizationClock(), true
	return decision, nil
}

func (session *identitySystemResourceAuthorization) require(
	action string,
) (uuid.UUID, error) {
	decision, err := session.decision(action)
	if err != nil {
		return uuid.Nil, err
	}
	permissionID, allowed := decision.PermissionID()
	if !allowed || permissionID == uuid.Nil {
		return uuid.Nil, publicresource.ErrForbidden
	}
	return permissionID, nil
}

func (session *identitySystemResourceAuthorization) allowed(
	action string,
) (bool, error) {
	decision, err := session.decision(action)
	if err != nil {
		return false, err
	}
	return decision.Allowed(), nil
}
