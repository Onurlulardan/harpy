package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) AddIdentifier(
	ctx context.Context, claims accesstoken.Claims, input LifecycleIdentifierAddInput,
) (LifecycleMutationView, error) {
	if !identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return LifecycleMutationView{}, ErrLifecycleHTTPInvalid
	}
	normalized, err := identifier.NormalizeEmailV1(input.Email)
	if err != nil {
		return LifecycleMutationView{}, ErrLifecycleHTTPInvalid
	}
	operationID := lifecycleOperationID(claims, "iam.identifier.add", input.IdempotencyKey)
	identifierID := lifecycleChildID(operationID, "identifier")
	requestSHA, err := lifecycleRequestSHA(struct{ Email string }{normalized.Normalized()})
	if err != nil {
		return LifecycleMutationView{}, err
	}
	var result lifecycle.MutationResult
	err = application.inCommand(ctx, claims, "iam.identifier.add", "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var executeErr error
			result, executeErr = application.runtime.service.AddIdentifier(scoped, scope,
				lifecycle.IdentifierAddCommand{
					Evidence: lifecycleEvidence(claims, permissionID, operationID, requestSHA),
					ID:       identifierID, UserID: claims.Subject(), Kind: "email_v1",
					DisplayValue: input.Email, NormalizedValue: normalized.Normalized(),
					NormalizedSHA256: normalized.NormalizedSHA256(),
				})
			return executeErr
		})
	if err != nil {
		return LifecycleMutationView{}, err
	}
	return LifecycleMutationView{ID: result.ID, Revision: result.Revision}, nil
}

func (application *lifecycleHTTPApplication) StartEmailVerification(
	ctx context.Context, claims accesstoken.Claims, input LifecycleResourceInput,
) error {
	if input.ID == uuid.Nil || !identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return ErrLifecycleHTTPInvalid
	}
	operationID := lifecycleOperationID(claims, "iam.email_verification.start", input.IdempotencyKey)
	requestSHA, err := lifecycleRequestSHA(struct{ IdentifierID uuid.UUID }{input.ID})
	if err != nil {
		return err
	}
	var target lifecycle.DeliveryTarget
	err = application.inSnapshot(ctx, claims, "read",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var readErr error
			target, readErr = application.runtime.service.EmailDeliveryTarget(scoped, scope,
				lifecycle.EmailDeliveryTargetQuery{IdentifierID: input.ID,
					Evidence: lifecycleEvidence(claims, permissionID, operationID, requestSHA)})
			return readErr
		})
	if err != nil {
		return err
	}
	tokenID, artifactID := lifecycleChildID(operationID, "token"), lifecycleChildID(operationID, "artifact")
	issued, token, err := lifecycle.NewCredentialGenerator().Generate(
		claims.Tenant(), lifecycle.PurposeEmailVerification, tokenID, application.oneTime)
	if err != nil {
		return ErrLifecycleHTTPUnavailable
	}
	defer issued.Destroy()
	delivery, err := application.deliveryArtifact(ctx, claims.Tenant(), artifactID,
		tokenID, lifecycle.PurposeEmailVerification, target, issued)
	if err != nil {
		return err
	}
	return application.inCommand(ctx, claims, "iam.email_verification.start", "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			_, executeErr := application.runtime.service.IssueEmailVerification(scoped, scope,
				lifecycle.EmailVerificationStartCommand{
					Evidence:     lifecycleEvidence(claims, permissionID, operationID, requestSHA),
					IdentifierID: input.ID, Token: token, Delivery: delivery,
				})
			return executeErr
		})
}
