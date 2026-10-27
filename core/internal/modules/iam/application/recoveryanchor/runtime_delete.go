package recoveryanchor

import (
	"context"

	anchorapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/recoveryanchor"
	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *Runtime) ProposeDeletion(
	ctx context.Context,
	input anchorapi.DeleteCommand,
) (anchorapi.Operation, error) {
	if runtime == nil || runtime.service == nil || ctx == nil || ctx.Err() != nil {
		return anchorapi.Operation{}, anchorapi.ErrInvalid
	}
	subject, err := parseRuntimeSubject(input.Subject)
	operationID, operationErr := uuid.Parse(input.OperationID)
	anchorID, anchorErr := uuid.Parse(input.AnchorID)
	digest, digestErr := parseRuntimeDigest(input.RequestSHA256)
	if err != nil || operationErr != nil || anchorErr != nil || digestErr != nil ||
		input.IdempotencyKey == "" {
		return anchorapi.Operation{}, anchorapi.ErrInvalid
	}
	request, err := runtime.scopeRequest(ctx, subject, "iam.recovery_anchor.delete", true)
	if err != nil {
		return anchorapi.Operation{}, err
	}
	command := anchordomain.DeleteCommand{Subject: subject,
		OperationID: operationID, AnchorID: anchorID,
		IdempotencyKey: input.IdempotencyKey, RequestSHA256: digest,
		OccurredAt: runtime.now()}
	var result anchordomain.ProposedMutation
	err = runtime.coordinator.InTenant(ctx, request,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			result, err = runtime.service.Delete(scoped, scope, command)
			return err
		})
	if err != nil {
		return anchorapi.Operation{}, mapRuntimeError(err)
	}
	return anchorapi.Operation{OperationID: runtimeAnchorUUID(result.OperationID),
		AnchorID: runtimeAnchorUUID(result.AnchorID), Kind: result.Kind, State: result.State}, nil
}
