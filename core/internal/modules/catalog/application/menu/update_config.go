package menu

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) UpdateConfig(
	ctx context.Context, scope platformtx.TenantScope, request ConfigWriteRequest,
) (ConfigDocument, error) {
	name, ok := normalizeConfigName(request.Name)
	if request.ID == uuid.Nil || request.ID.Version() != 7 || !ok ||
		!validConfigJSON(request.Config) || !idempotencyPattern.MatchString(request.IdempotencyKey) {
		return ConfigDocument{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, operationUpdate, permissionManage)
	if err != nil {
		return ConfigDocument{}, err
	}
	now := application.clock.Now()
	if !canonicalTime(now) {
		return ConfigDocument{}, ErrUnavailable
	}
	document, err := store.UpdateConfig(ctx, ConfigMutation{
		ID: request.ID, Name: name, Config: request.Config, Active: request.Active,
		ActorID: identity.ActorID, OccurredAt: now,
	})
	if err != nil || document.ID == uuid.Nil {
		return ConfigDocument{}, errUnavailableOr(err)
	}
	return document, nil
}
