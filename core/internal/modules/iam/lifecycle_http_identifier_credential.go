package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *lifecycleHTTPApplication) VerifyIdentifier(
	ctx context.Context, input LifecycleCredentialInput,
) error {
	return application.completeCredential(ctx, input, lifecycle.PurposeEmailVerification,
		"iam.identifier.verify", nil, func(scoped context.Context,
			scope platformtx.CredentialTenantScope, record lifecycle.CredentialRecord,
			proof lifecycle.CredentialProof, evidence lifecycle.CredentialEvidence) error {
			_, err := application.runtime.service.VerifyIdentifier(scoped, scope,
				lifecycle.IdentifierVerifyCommand{Evidence: evidence, Proof: proof})
			return err
		})
}

type lifecycleCredentialAction func(context.Context, platformtx.CredentialTenantScope,
	lifecycle.CredentialRecord, lifecycle.CredentialProof, lifecycle.CredentialEvidence) error

func (application *lifecycleHTTPApplication) validateCredential(
	ctx context.Context,
	input LifecycleCredentialInput,
	purpose lifecycle.CredentialPurpose,
	operation string,
) error {
	input.Token = append([]byte(nil), input.Token...)
	return application.completeCredential(
		ctx,
		input,
		purpose,
		operation,
		nil,
		func(
			context.Context,
			platformtx.CredentialTenantScope,
			lifecycle.CredentialRecord,
			lifecycle.CredentialProof,
			lifecycle.CredentialEvidence,
		) error {
			return nil
		},
	)
}

func (application *lifecycleHTTPApplication) completeCredential(
	ctx context.Context, input LifecycleCredentialInput,
	purpose lifecycle.CredentialPurpose, operation string,
	requestPayload []byte,
	execute lifecycleCredentialAction,
) error {
	defer clear(input.Token)
	if application == nil || ctx == nil || ctx.Err() != nil || input.TenantSlug == "" ||
		len(input.Token) == 0 || execute == nil {
		return ErrLifecycleHTTPInvalidCredential
	}
	presented, err := lifecycle.ParseCredential(input.Token)
	if err != nil {
		return ErrLifecycleHTTPInvalidCredential
	}
	defer presented.Destroy()
	tenantID, err := application.pool.ResolveActiveTenantIDBySlug(ctx, input.TenantSlug)
	if err != nil {
		return ErrLifecycleHTTPInvalidCredential
	}
	selector, err := presented.Selector()
	if err != nil {
		return ErrLifecycleHTTPInvalidCredential
	}
	binding, err := lifecycle.CredentialBindingSHA256(tenantID, purpose, selector)
	if err != nil {
		return ErrLifecycleHTTPInvalidCredential
	}
	txPurpose, valid := purposeForScope(purpose)
	if !valid {
		return ErrLifecycleHTTPInvalidCredential
	}
	lookup, err := platformtx.NewTenantRouteCredentialLookup(operation, txPurpose, tenantID, binding)
	if err != nil {
		return ErrLifecycleHTTPUnavailable
	}
	err = application.coordinator.InCredentialTenant(ctx,
		platformtx.CredentialScopeRequest{Lookup: lookup, Resolver: application.resolver},
		func(scoped context.Context, scope platformtx.CredentialTenantScope) error {
			record, lookupErr := application.runtime.service.LookupCredential(scoped, scope,
				lifecycle.CredentialLookup{Purpose: purpose, Selector: selector, BindingSHA256: binding})
			if lookupErr != nil {
				return lookupErr
			}
			proof, proofErr := presented.Proof(tenantID, record, application.oneTime)
			if proofErr != nil {
				return lifecycle.ErrDenied
			}
			requestSHA, requestErr := lifecycleRequestSHA(
				struct{ TokenID uuid.UUID }{record.TokenID},
			)
			if requestPayload != nil {
				requestSHA, requestErr = presented.RequestSHA256(
					tenantID, purpose, record.TokenID, requestPayload,
				)
			}
			if requestErr != nil {
				return lifecycle.ErrDenied
			}
			operationID := identityDeterministicUUID([]byte("harpy/iam/lifecycle-credential/v1"),
				tenantID[:], record.TokenID[:], []byte(operation))
			return execute(scoped, scope, record, proof,
				lifecycle.CredentialEvidence{OperationID: operationID, RequestSHA256: requestSHA})
		})
	if err != nil {
		return ErrLifecycleHTTPInvalidCredential
	}
	return nil
}

func purposeForScope(value lifecycle.CredentialPurpose) (platformtx.CredentialPurpose, bool) {
	switch value {
	case lifecycle.PurposeEmailVerification:
		return platformtx.CredentialEmailVerification, true
	case lifecycle.PurposePasswordReset:
		return platformtx.CredentialPasswordReset, true
	case lifecycle.PurposeInitialAdminSetup:
		return platformtx.CredentialInitialAdminSetup, true
	case lifecycle.PurposePortalSetup:
		return platformtx.CredentialPortalSetup, true
	default:
		return "", false
	}
}
