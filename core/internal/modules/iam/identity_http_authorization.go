package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) authorize(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityEvidenceSource,
	resource IdentityResource,
	action string,
) (uuid.UUID, error) {
	resourceKey, err := identityResourceKey(resource)
	if err != nil {
		return uuid.Nil, ErrIdentityHTTPInvalid
	}
	request, err := authorization.NewRequest(
		resourceKey, action, evidence.TokenVersion(), evidence.AuthorizationVersion(),
	)
	if err != nil {
		return uuid.Nil, ErrIdentityHTTPForbidden
	}
	decision, err := application.authorizer.Authorize(ctx, scope, request)
	if err != nil || decision.Validate() != nil {
		return uuid.Nil, ErrIdentityHTTPUnavailable
	}
	if !decision.Allowed() {
		return uuid.Nil, ErrIdentityHTTPForbidden
	}
	permissionID, valid := decision.PermissionID()
	if !valid || permissionID == uuid.Nil {
		return uuid.Nil, ErrIdentityHTTPUnavailable
	}
	return permissionID, nil
}

type identityEvidenceSource interface {
	Subject() uuid.UUID
	Tenant() uuid.UUID
	TokenVersion() int64
	AuthorizationVersion() int64
}

func identityQueryEvidence(
	claims identityEvidenceSource,
	permissionID uuid.UUID,
) identityfacade.QueryEvidence {
	return identityfacade.QueryEvidence{
		ActorID:                      identityfacade.IdentityID(claims.Subject()),
		PermissionID:                 identityfacade.IdentityID(permissionID),
		ExpectedTokenVersion:         claims.TokenVersion(),
		ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
	}
}

func identityMutationEvidence(
	claims identityEvidenceSource,
	permissionID uuid.UUID,
	operationID uuid.UUID,
	requestSHA [32]byte,
) identityfacade.MutationEvidence {
	return identityfacade.MutationEvidence{
		OperationID:                  identityfacade.IdentityID(operationID),
		RequestSHA256:                requestSHA,
		ActorID:                      identityfacade.IdentityID(claims.Subject()),
		PermissionID:                 identityfacade.IdentityID(permissionID),
		ExpectedTokenVersion:         claims.TokenVersion(),
		ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
	}
}
