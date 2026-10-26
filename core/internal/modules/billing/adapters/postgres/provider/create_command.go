package provider

import (
	"context"
	"fmt"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

func (storage *commandStore) CreateCommand(
	ctx context.Context,
	mutation providerapp.CommandMutation,
) (providerapp.CommandSnapshot, error) {
	if !storage.usable(ctx) || mutation.TenantID != storage.tenantID ||
		mutation.ActorID != storage.actorID {
		return providerapp.CommandSnapshot{}, providerapp.ErrCommandRequest
	}
	var authorityID uuid.UUID
	if err := storage.database.queryRow(
		ctx, resolveCommandAuthoritySQL, storage.tenantID,
	).Scan(&authorityID); err != nil || authorityID == uuid.Nil {
		return providerapp.CommandSnapshot{}, fmt.Errorf(
			"resolve billing provider command authority: %w", err,
		)
	}
	result, err := scanCommandSnapshot(storage.database.queryRow(
		ctx, createCommandSQL, authorityID, mutation.TenantID,
		mutation.CommandID, mutation.FinalizerJobID, mutation.ActorID,
		mutation.ProviderObjectMappingID, mutation.SubscriptionID,
		string(mutation.Kind), mutation.EffectKey, mutation.RequestSHA256[:],
		mutation.RequestedAt,
	))
	if err != nil {
		return providerapp.CommandSnapshot{}, fmt.Errorf(
			"persist billing provider command: %w", classifyCommandError(err),
		)
	}
	if err := storage.effects.EnqueueCommandRoute(ctx, result); err != nil {
		return providerapp.CommandSnapshot{}, fmt.Errorf(
			"enqueue billing provider command route: %w", err,
		)
	}
	return result, nil
}
