package work

import (
	"context"
	"strings"

	workpostgres "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type ProjectionApplyInput struct {
	ActorID, TenantID uuid.UUID
	Item              ports.ProjectionItem
	IdempotencyKey    string
}

type ProjectionRepairInput struct {
	ActorID, TenantID uuid.UUID
	Owner             ports.SourceOwner
	Kind              string
	SourceID          ports.ID
	Through           int64
	Items             []ports.ProjectionItem
	IdempotencyKey    string
}

func (runtime *PlatformHTTPRuntime) ApplyProjection(ctx context.Context,
	input ProjectionApplyInput,
) (receipt ports.OperationReceipt, err error) {
	if runtime == nil || runtime.projections == nil || input.ActorID == uuid.Nil ||
		input.TenantID == uuid.Nil || strings.TrimSpace(input.IdempotencyKey) == "" {
		return receipt, ErrHTTPInvalid
	}
	err = runtime.projectionCommand(ctx, input.ActorID, input.TenantID,
		"work.projection.apply", func(operation context.Context, scope platformtx.TenantScope) error {
			var applyErr error
			receipt, applyErr = runtime.projections.ConsumeProjectionEvent(operation,
				workpostgres.BridgeScope(scope), ports.Mutation{
					IdempotencyKey: input.IdempotencyKey}, input.Item)
			return applyErr
		})
	return receipt, err
}

func (runtime *PlatformHTTPRuntime) RepairProjection(ctx context.Context,
	input ProjectionRepairInput,
) (receipt ports.OperationReceipt, err error) {
	if runtime == nil || runtime.projections == nil || input.ActorID == uuid.Nil ||
		input.TenantID == uuid.Nil || strings.TrimSpace(input.IdempotencyKey) == "" {
		return receipt, ErrHTTPInvalid
	}
	err = runtime.projectionCommand(ctx, input.ActorID, input.TenantID,
		"work.projection.repair", func(operation context.Context, scope platformtx.TenantScope) error {
			var repairErr error
			receipt, repairErr = runtime.projections.RepairProjection(operation,
				workpostgres.BridgeScope(scope), ports.Mutation{
					IdempotencyKey: input.IdempotencyKey}, input.Owner, input.Kind,
				input.SourceID, input.Through, input.Items)
			return repairErr
		})
	return receipt, err
}

func (runtime *PlatformHTTPRuntime) projectionCommand(ctx context.Context,
	actorID, tenantID uuid.UUID, operation string,
	action func(context.Context, platformtx.TenantScope) error,
) error {
	group, generation, err := runtime.resolver.ResolveTenantWorkGenerations(ctx, tenantID)
	if err != nil || group < 1 || generation < 1 {
		return ErrHTTPUnavailable
	}
	request, err := platformtx.AdmitCommandScope(actorID, tenantID, operation,
		"work", group, map[string]int64{"work": generation})
	if err != nil {
		return ErrHTTPUnavailable
	}
	return runtime.scopes.InTenant(ctx, request, action)
}
