package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) TombstoneIdentifier(
	ctx context.Context, claims accesstoken.Claims, input LifecycleRevisionInput,
) error {
	return application.identifierRevision(ctx, claims, input,
		"iam.identifier.tombstone", application.runtime.service.TombstoneIdentifier)
}

func (application *lifecycleHTTPApplication) RetireIdentifier(
	ctx context.Context, claims accesstoken.Claims, input LifecycleRevisionInput,
) error {
	return application.identifierRevision(ctx, claims, input,
		"iam.identifier.retire", application.runtime.service.ArchiveIdentifier)
}

type lifecycleRevisionOperation func(context.Context, platformtx.TenantScope,
	lifecycle.RevisionCommand) (lifecycle.MutationResult, error)

func (application *lifecycleHTTPApplication) identifierRevision(
	ctx context.Context, claims accesstoken.Claims, input LifecycleRevisionInput,
	operation string, execute lifecycleRevisionOperation,
) error {
	if input.ID == uuid.Nil || input.ExpectedRevision < 1 ||
		!identityIdempotencyPattern.MatchString(input.IdempotencyKey) || execute == nil {
		return ErrLifecycleHTTPInvalid
	}
	operationID := lifecycleOperationID(claims, operation, input.IdempotencyKey)
	requestSHA, err := lifecycleRequestSHA(struct {
		ID       uuid.UUID
		Revision int64
	}{
		input.ID, input.ExpectedRevision,
	})
	if err != nil {
		return err
	}
	return application.inCommand(ctx, claims, operation, "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			_, executeErr := execute(scoped, scope, lifecycle.RevisionCommand{
				Evidence: lifecycleEvidence(claims, permissionID, operationID, requestSHA),
				ID:       input.ID, ExpectedRevision: input.ExpectedRevision,
			})
			return executeErr
		})
}
