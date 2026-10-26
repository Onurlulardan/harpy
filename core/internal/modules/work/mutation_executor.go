package work

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type MutationDependencies struct {
	UnitOfWork ports.UnitOfWork
	Authorizer ports.Authorizer
	Events     ports.EventPublisher
	Clock      ports.Clock
	IDs        ports.IDGenerator
}

type MutationExecutor struct{ dependencies MutationDependencies }

func NewMutationExecutor(dependencies MutationDependencies) (*MutationExecutor, error) {
	if dependencies.UnitOfWork == nil || dependencies.Authorizer == nil ||
		dependencies.Events == nil || dependencies.Clock == nil || dependencies.IDs == nil {
		return nil, ports.ErrUnavailable
	}
	return &MutationExecutor{dependencies: dependencies}, nil
}

func (executor *MutationExecutor) Mutate(
	ctx context.Context, scope ports.Scope, authority string, resource ports.ID,
	meta ports.Mutation, hash [32]byte, action ports.MutationAction,
) (ports.OperationReceipt, error) {
	if executor == nil || ctx == nil || ctx.Err() != nil || scope == nil || action == nil ||
		strings.TrimSpace(meta.IdempotencyKey) == "" || len(meta.IdempotencyKey) > 200 {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	if err := executor.dependencies.Authorizer.Authorize(ctx, scope, authority, resource); err != nil {
		return ports.OperationReceipt{}, ports.ErrForbidden
	}
	stores, err := executor.dependencies.UnitOfWork.Bind(ctx, scope)
	if err != nil || stores == nil || stores.Operations() == nil {
		return ports.OperationReceipt{}, ports.ErrUnavailable
	}
	operationID, err := executor.dependencies.IDs.New(executor.dependencies.Clock.Now().UTC())
	if err != nil {
		return ports.OperationReceipt{}, ports.ErrUnavailable
	}
	receipt, duplicate, err := stores.Operations().Acquire(ctx, operationID,
		meta.IdempotencyKey, authority, hash)
	if err != nil || duplicate {
		return receipt, err
	}
	receipt, payload, err := action(ctx, stores)
	if err != nil {
		return ports.OperationReceipt{}, err
	}
	if err = executor.finalize(ctx, scope, stores, operationID, authority,
		meta.IdempotencyKey, receipt, payload); err != nil {
		return ports.OperationReceipt{}, err
	}
	return receipt, nil
}

func (executor *MutationExecutor) finalize(ctx context.Context, scope ports.Scope, stores ports.Stores,
	operationID ports.ID, authority, idempotencyKey string,
	receipt ports.OperationReceipt, payload []byte,
) error {
	now := executor.dependencies.Clock.Now().UTC()
	auditID, err := executor.dependencies.IDs.New(now)
	if err != nil {
		return ports.ErrUnavailable
	}
	sequence, err := stores.Operations().AppendAudit(ctx, auditID, operationID, authority,
		scope.ActorID(), receipt.AggregateID, receipt.RowVersion)
	if err != nil {
		return err
	}
	intent := ports.EventIntent{
		EventID: operationID, Type: authority, AggregateKind: aggregateKind(authority),
		AggregateID: receipt.AggregateID, AggregateVersion: receipt.RowVersion,
		EventSequence: sequence, OccurredAt: now, IdempotencyKey: idempotencyKey,
		Payload: append([]byte{}, payload...),
	}
	if err = executor.dependencies.Events.Record(ctx, scope, intent); err != nil {
		return err
	}
	return stores.Operations().PersistResult(ctx, operationID, receipt, payload)
}

func aggregateKind(authority string) string {
	parts := strings.Split(authority, ".")
	if len(parts) >= 3 && parts[0] == "work" {
		return "work." + parts[1]
	}
	return "work.aggregate"
}

var _ ports.Mutator = (*MutationExecutor)(nil)
