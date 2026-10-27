package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) ListSessions(
	ctx context.Context, claims accesstoken.Claims, input LifecyclePageInput,
) ([]lifecycle.Session, error) {
	if input.Limit < 1 || input.Limit > 100 || input.AfterTime.IsZero() != (input.AfterID == uuid.Nil) {
		return nil, ErrLifecycleHTTPInvalid
	}
	var result []lifecycle.Session
	err := application.inSnapshot(ctx, claims, "read",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var readErr error
			result, readErr = application.runtime.service.GetSessions(scoped, scope,
				lifecycle.SessionListQuery{UserID: claims.Subject(), ActorID: claims.Subject(),
					PermissionID: permissionID, ExpectedTokenVersion: claims.TokenVersion(),
					ExpectedAuthorizationVersion: claims.AuthorizationVersion(), Limit: input.Limit,
					AfterLastSeenAt: input.AfterTime, AfterID: input.AfterID})
			return readErr
		})
	return result, err
}

func (application *lifecycleHTTPApplication) RevokeSession(
	ctx context.Context, claims accesstoken.Claims, input LifecycleRevisionInput,
) error {
	return application.sessionRevision(ctx, claims, input,
		"iam.session.revoke_one", application.runtime.service.RevokeSession)
}

func (application *lifecycleHTTPApplication) sessionRevision(
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

func (application *lifecycleHTTPApplication) RevokeAllSessions(
	ctx context.Context, claims accesstoken.Claims, idempotencyKey string,
) error {
	if !identityIdempotencyPattern.MatchString(idempotencyKey) {
		return ErrLifecycleHTTPInvalid
	}
	operationID := lifecycleOperationID(claims, "iam.session.revoke_all", idempotencyKey)
	requestSHA, _ := lifecycleRequestSHA(struct{ UserID uuid.UUID }{claims.Subject()})
	return application.inCommand(ctx, claims, "iam.session.revoke_all", "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			_, executeErr := application.runtime.service.PurgeSessions(scoped, scope,
				lifecycle.SessionRevokeAllCommand{Evidence: lifecycleEvidence(claims, permissionID, operationID, requestSHA)})
			return executeErr
		})
}
