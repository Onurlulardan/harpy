package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) ListAuthenticators(
	ctx context.Context, claims accesstoken.Claims, input LifecyclePageInput,
) ([]lifecycle.Authenticator, error) {
	if input.Limit < 1 || input.Limit > 100 || input.AfterTime.IsZero() != (input.AfterID == uuid.Nil) {
		return nil, ErrLifecycleHTTPInvalid
	}
	var result []lifecycle.Authenticator
	err := application.inSnapshot(ctx, claims, "read",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var readErr error
			result, readErr = application.runtime.service.ListAuthenticators(scoped, scope,
				lifecycle.AuthenticatorListQuery{UserID: claims.Subject(), ActorID: claims.Subject(),
					PermissionID: permissionID, ExpectedTokenVersion: claims.TokenVersion(),
					ExpectedAuthorizationVersion: claims.AuthorizationVersion(), Limit: input.Limit,
					AfterUpdatedAt: input.AfterTime, AfterID: input.AfterID})
			return readErr
		})
	return result, err
}

func (application *lifecycleHTTPApplication) SuspendAuthenticator(
	ctx context.Context, claims accesstoken.Claims, input LifecycleRevisionInput,
) error {
	return application.authenticatorRevision(ctx, claims, input,
		"iam.authenticator.suspend", application.runtime.service.SuspendAuthenticator)
}

func (application *lifecycleHTTPApplication) RetireAuthenticator(
	ctx context.Context, claims accesstoken.Claims, input LifecycleRevisionInput,
) error {
	return application.authenticatorRevision(ctx, claims, input,
		"iam.authenticator.retire", application.runtime.service.RetireAuthenticator)
}

func (application *lifecycleHTTPApplication) authenticatorRevision(
	ctx context.Context, claims accesstoken.Claims, input LifecycleRevisionInput,
	operation string, execute lifecycleRevisionOperation,
) error {
	if input.ID == uuid.Nil || input.ExpectedRevision < 1 ||
		!identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return ErrLifecycleHTTPInvalid
	}
	operationID := lifecycleOperationID(claims, operation, input.IdempotencyKey)
	requestSHA, _ := lifecycleRequestSHA(struct {
		ID       uuid.UUID
		Revision int64
	}{input.ID, input.ExpectedRevision})
	return application.inCommand(ctx, claims, operation, "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			_, executeErr := execute(scoped, scope, lifecycle.RevisionCommand{
				Evidence: lifecycleEvidence(claims, permissionID, operationID, requestSHA),
				ID:       input.ID, ExpectedRevision: input.ExpectedRevision})
			return executeErr
		})
}
