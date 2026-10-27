package providerapp

import (
	"context"
	"fmt"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (commands *CommandCommands) CreateCommand(
	ctx context.Context,
	scope platformtx.TenantScope,
	request SubmitCommandRequest,
) (CommandSnapshot, error) {
	if commands == nil || commands.store == nil || commands.nextID == nil ||
		commands.clock == nil || ctx == nil || ctx.Err() != nil || !request.valid() ||
		!platformtx.ContextOwnsScope(ctx, scope) ||
		platformtx.RequireActive(scope, platformtx.Command) != nil ||
		scope.TenantID() == uuid.Nil || scope.ActorID() == uuid.Nil {
		return CommandSnapshot{}, ErrCommandRequest
	}
	commandID, err := nextCommandID(commands.nextID)
	if err != nil {
		return CommandSnapshot{}, err
	}
	finalizerID, err := nextCommandID(commands.nextID)
	if err != nil || finalizerID == commandID {
		return CommandSnapshot{}, ErrConfiguration
	}
	requestedAt, err := canonicalInstant(commands.clock())
	if err != nil {
		return CommandSnapshot{}, ErrConfiguration
	}
	result, err := commands.store.CreateCommand(ctx, CommandMutation{
		CommandID: commandID, FinalizerJobID: finalizerID,
		TenantID: scope.TenantID(), ActorID: scope.ActorID(),
		ProviderObjectMappingID: request.ProviderObjectMappingID,
		SubscriptionID:          request.SubscriptionID, Kind: request.Kind,
		EffectKey: request.EffectKey, RequestSHA256: request.RequestSHA256,
		RequestedAt: requestedAt,
	})
	if err != nil {
		return CommandSnapshot{}, fmt.Errorf("create billing provider command: %w", err)
	}
	if !result.valid() || result.TenantID != scope.TenantID() ||
		result.ActorID != scope.ActorID() || result.SubscriptionID != request.SubscriptionID ||
		result.Kind != request.Kind || result.EffectKey != request.EffectKey ||
		result.RequestSHA256 != request.RequestSHA256 || result.State != CommandQueued ||
		result.Revision != 1 || (!result.Replay &&
		(result.CommandID != commandID || result.FinalizerJobID != finalizerID)) {
		return CommandSnapshot{}, ErrCommandState
	}
	return result, nil
}

func nextCommandID(source CommandIDSource) (uuid.UUID, error) {
	identifier, err := source()
	if err != nil || identifier == uuid.Nil || identifier.Version() != uuid.Version(7) {
		return uuid.Nil, ErrConfiguration
	}
	return identifier, nil
}
