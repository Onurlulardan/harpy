package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) StartPortalSetup(
	ctx context.Context, claims accesstoken.Claims, input LifecycleResourceInput,
) error {
	if input.ID == uuid.Nil || !identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return ErrLifecycleHTTPInvalid
	}
	operationID := lifecycleOperationID(claims, "iam.portal_setup.start", input.IdempotencyKey)
	requestSHA, _ := lifecycleRequestSHA(struct{ BindingID uuid.UUID }{input.ID})
	var target lifecycle.DeliveryTarget
	err := application.inSnapshot(ctx, claims, "read",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			var readErr error
			target, readErr = application.runtime.service.PortalDeliveryTarget(scoped, scope,
				lifecycle.PortalDeliveryTargetQuery{BindingID: input.ID,
					Evidence: lifecycleEvidence(claims, permissionID, operationID, requestSHA)})
			return readErr
		})
	if err != nil {
		return err
	}
	tokenID, artifactID := lifecycleChildID(operationID, "token"), lifecycleChildID(operationID, "artifact")
	issued, token, err := lifecycle.NewCredentialGenerator().Generate(
		claims.Tenant(), lifecycle.PurposePortalSetup, tokenID, application.oneTime)
	if err != nil {
		return ErrLifecycleHTTPUnavailable
	}
	defer issued.Destroy()
	delivery, err := application.deliveryArtifact(ctx, claims.Tenant(), artifactID,
		tokenID, lifecycle.PurposePortalSetup, target, issued)
	if err != nil {
		return err
	}
	return application.inCommand(ctx, claims, "iam.portal_setup.start", "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			_, executeErr := application.runtime.service.CreatePortalSetup(scoped, scope,
				lifecycle.PortalSetupStartCommand{
					Evidence:  lifecycleEvidence(claims, permissionID, operationID, requestSHA),
					BindingID: input.ID, Token: token, Delivery: delivery,
				})
			return executeErr
		})
}

func (application *lifecycleHTTPApplication) CompletePortalSetup(
	ctx context.Context, input LifecyclePasswordResetCompleteInput,
) error {
	defer clear(input.Password)
	defer clear(input.Token)
	credentialInput := LifecycleCredentialInput{
		TenantSlug: input.TenantSlug,
		Token:      input.Token,
	}
	if err := application.validateCredential(
		ctx, credentialInput, lifecycle.PurposePortalSetup,
		"iam.portal_setup.complete",
	); err != nil {
		return err
	}
	requestPayload := append([]byte(nil), input.Password...)
	defer clear(requestPayload)
	credential, err := application.passwordCredential(ctx, input.Password,
		[]string{input.TenantSlug})
	if err != nil {
		return err
	}
	return application.completeCredential(ctx, credentialInput,
		lifecycle.PurposePortalSetup, "iam.portal_setup.complete", requestPayload,
		func(scoped context.Context, scope platformtx.CredentialTenantScope,
			_ lifecycle.CredentialRecord, proof lifecycle.CredentialProof,
			evidence lifecycle.CredentialEvidence) error {
			_, executeErr := application.runtime.service.RedeemPortalSetup(scoped, scope,
				lifecycle.PortalSetupCompleteCommand{Evidence: evidence, Proof: proof,
					Salt: credential.Salt(), Tag: credential.Tag()})
			return executeErr
		})
}

func (application *lifecycleHTTPApplication) DecidePortalSetup(
	ctx context.Context, claims accesstoken.Claims, input LifecyclePortalDecisionInput,
) error {
	decision, valid := canonicalPortalApprovalDecision(input.Decision)
	if !valid {
		return ErrLifecycleHTTPInvalid
	}
	base := input.LifecycleRevisionInput
	if base.ID == uuid.Nil || base.ExpectedRevision < 1 ||
		!identityIdempotencyPattern.MatchString(base.IdempotencyKey) {
		return ErrLifecycleHTTPInvalid
	}
	operationID := lifecycleOperationID(claims, "iam.portal_approval.decide", base.IdempotencyKey)
	requestSHA, _ := lifecycleRequestSHA(struct {
		ID       uuid.UUID
		Revision int64
		Decision string
	}{
		base.ID, base.ExpectedRevision, input.Decision})
	return application.inCommand(ctx, claims, "iam.portal_approval.decide", "update",
		func(scoped context.Context, scope platformtx.TenantScope, permissionID uuid.UUID) error {
			_, executeErr := application.runtime.service.DecidePortalApproval(scoped, scope,
				lifecycle.PortalDecisionCommand{
					Evidence:  lifecycleEvidence(claims, permissionID, operationID, requestSHA),
					BindingID: base.ID, ExpectedRevision: base.ExpectedRevision,
					Decision: decision,
				})
			return executeErr
		})
}

func canonicalPortalApprovalDecision(
	value string,
) (lifecycle.PortalApprovalDecision, bool) {
	switch value {
	case "approve":
		return lifecycle.PortalApprovalApproved, true
	case "reject":
		return lifecycle.PortalApprovalRejected, true
	default:
		return "", false
	}
}
