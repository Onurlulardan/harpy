package menu

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) CreateConfig(
	ctx context.Context, scope platformtx.TenantScope, request ConfigWriteRequest,
) (ConfigDocument, error) {
	name, ok := normalizeConfigName(request.Name)
	if request.ID != uuid.Nil || !ok ||
		!validConfigJSON(request.Config) || !idempotencyPattern.MatchString(request.IdempotencyKey) {
		return ConfigDocument{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand, operationUpdate, permissionManage)
	if err != nil {
		return ConfigDocument{}, err
	}
	now := application.clock.Now()
	id, idErr := application.ids.New()
	if !canonicalTime(now) || idErr != nil || id == uuid.Nil || id.Version() != 7 {
		return ConfigDocument{}, ErrUnavailable
	}
	document, err := store.CreateConfig(ctx, ConfigMutation{
		ID: id, Name: name, Config: request.Config, Active: true,
		ActorID: identity.ActorID, OccurredAt: now,
	})
	if err != nil || document.ID == uuid.Nil {
		return ConfigDocument{}, errUnavailableOr(err)
	}
	return document, nil
}

func errUnavailableOr(err error) error {
	if err != nil {
		return err
	}
	return ErrUnavailable
}
