package iam

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *lifecycleHTTPApplication) StartPasswordReset(
	ctx context.Context, input LifecyclePasswordResetStartInput,
) error {
	if application == nil || ctx == nil || ctx.Err() != nil || input.TenantSlug == "" ||
		!identityIdempotencyPattern.MatchString(input.IdempotencyKey) {
		return ErrLifecycleHTTPInvalid
	}
	value, kind, txKind, err := normalizeResetIdentifier(input.IdentifierKind, input.Identifier)
	if err != nil {
		return nil
	}
	tenantID, err := application.pool.ResolveActiveTenantIDBySlug(ctx, input.TenantSlug)
	if err != nil {
		return nil
	}
	directoryHMAC := invitationDirectoryHMAC(application.directory.material, value.Normalized())
	defer clear(directoryHMAC[:])
	lookup, err := platformtx.NewPasswordResetCredentialLookup(
		"iam.password_reset.start", tenantID, txKind, application.directory.id, directoryHMAC)
	if err != nil {
		return ErrLifecycleHTTPUnavailable
	}
	var target lifecycle.DeliveryTarget
	err = application.coordinator.InCredentialTenant(ctx,
		platformtx.CredentialScopeRequest{Lookup: lookup, Resolver: application.resolver},
		func(scoped context.Context, scope platformtx.CredentialTenantScope) error {
			var readErr error
			target, readErr = application.runtime.service.PasswordDeliveryTarget(scoped, scope,
				lifecycle.PasswordDeliveryTargetQuery{IdentifierKind: kind,
					IdentifierSHA256: value.NormalizedSHA256()})
			return readErr
		})
	if err != nil {
		return nil
	}
	identifierSHA := value.NormalizedSHA256()
	operationID := identityDeterministicUUID(
		[]byte("harpy/iam/password-reset-start/v1"),
		tenantID[:], identifierSHA[:], []byte(input.IdempotencyKey),
	)
	tokenID, artifactID := lifecycleChildID(operationID, "token"), lifecycleChildID(operationID, "artifact")
	requestSHA, _ := lifecycleRequestSHA(struct {
		Kind   string
		Digest [32]byte
	}{kind, identifierSHA})
	issued, token, err := lifecycle.NewCredentialGenerator().Generate(
		tenantID, lifecycle.PurposePasswordReset, tokenID, application.oneTime)
	if err != nil {
		return ErrLifecycleHTTPUnavailable
	}
	defer issued.Destroy()
	delivery, err := application.deliveryArtifact(ctx, tenantID, artifactID, tokenID,
		lifecycle.PurposePasswordReset, target, issued)
	if err != nil {
		return err
	}
	err = application.coordinator.InCredentialTenant(ctx,
		platformtx.CredentialScopeRequest{Lookup: lookup, Resolver: application.resolver},
		func(scoped context.Context, scope platformtx.CredentialTenantScope) error {
			_, executeErr := application.runtime.service.StartPasswordReset(scoped, scope,
				lifecycle.PasswordResetStartCommand{
					Evidence:       lifecycle.CredentialEvidence{OperationID: operationID, RequestSHA256: requestSHA},
					IdentifierKind: kind, IdentifierSHA256: value.NormalizedSHA256(),
					Token: token, Delivery: delivery,
				})
			return executeErr
		})
	if err != nil {
		return ErrLifecycleHTTPUnavailable
	}
	return nil
}

func normalizeResetIdentifier(kind, raw string) (identifier.Value, string,
	platformtx.CredentialIdentifierKind, error) {
	switch kind {
	case "email_v1":
		value, err := identifier.NormalizeEmailV1(raw)
		return value, kind, platformtx.CredentialEmailIdentifier, err
	case "portal_v1":
		value, err := identifier.NormalizePortalV1(raw)
		return value, kind, platformtx.CredentialPortalIdentifier, err
	default:
		return identifier.Value{}, "", "", ErrLifecycleHTTPInvalid
	}
}
