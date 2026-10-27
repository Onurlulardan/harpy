package recoveryanchor

import (
	"context"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *Runtime) ProposeRegistration(
	ctx context.Context,
	input anchorapi.RegisterCommand,
) (anchorapi.Operation, error) {
	if runtime == nil || runtime.service == nil || ctx == nil || ctx.Err() != nil {
		return anchorapi.Operation{}, anchorapi.ErrInvalid
	}
	subject, err := parseRuntimeSubject(input.Subject)
	operationID, operationErr := uuid.Parse(input.OperationID)
	digest, digestErr := parseRuntimeDigest(input.RequestSHA256)
	if err != nil || operationErr != nil || digestErr != nil ||
		input.IdempotencyKey == "" || len(input.CredentialJSON) < 2 {
		return anchorapi.Operation{}, anchorapi.ErrInvalid
	}
	lookup, err := runtime.loadRegistration(ctx, subject, input.OperationID)
	if err != nil {
		return anchorapi.Operation{}, err
	}
	evidence, err := runtime.verifier.Verify(ctx,
		input.CredentialJSON, anchorapi.RegistrationExpectation{
			Challenge: lookup.Challenge, RPID: lookup.RPID,
			AllowedOrigins: lookup.Origins})
	if err != nil {
		return anchorapi.Operation{}, err
	}
	aaguid, err := uuid.Parse(evidence.AAGUID)
	if err != nil {
		return anchorapi.Operation{}, anchorapi.ErrInvalid
	}
	command := anchordomain.RegisterCommand{Subject: subject,
		OperationID: operationID, IdempotencyKey: input.IdempotencyKey,
		RequestSHA256: digest, RPID: lookup.RPID,
		Origins: lookup.Origins, OriginsSHA256: lookup.OriginsSHA256,
		OccurredAt: runtime.now(), Evidence: anchordomain.RegistrationEvidence{
			CredentialID: evidence.CredentialID, PublicKeyCOSE: evidence.PublicKeyCOSE,
			AAGUID: aaguid, AttestationFormat: evidence.AttestationFormat,
			AttestationType: evidence.AttestationType, SignCount: evidence.SignCount}}
	return runtime.register(ctx, subject, command)
}

func (runtime *Runtime) loadRegistration(
	ctx context.Context, subject anchordomain.Subject, operationID string,
) (anchordomain.RegistrationLookup, error) {
	request, err := runtime.scopeRequest(ctx, subject, "iam.recovery_anchor.registration", false)
	if err != nil {
		return anchordomain.RegistrationLookup{}, err
	}
	var result anchordomain.RegistrationLookup
	err = runtime.coordinator.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.service.LoadRegistration(scoped, scope, subject, operationID)
			return err
		})
	return result, mapRuntimeError(err)
}

func (runtime *Runtime) register(
	ctx context.Context, subject anchordomain.Subject, command anchordomain.RegisterCommand,
) (anchorapi.Operation, error) {
	request, err := runtime.scopeRequest(ctx, subject, "iam.recovery_anchor.register", true)
	if err != nil {
		return anchorapi.Operation{}, err
	}
	var result anchordomain.ProposedMutation
	err = runtime.coordinator.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.service.Register(scoped, scope, command)
			return err
		})
	if err != nil {
		return anchorapi.Operation{}, mapRuntimeError(err)
	}
	return anchorapi.Operation{OperationID: runtimeAnchorUUID(result.OperationID),
		AnchorID: runtimeAnchorUUID(result.AnchorID), Kind: result.Kind, State: result.State}, nil
}
