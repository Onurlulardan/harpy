package menu

import (
	"context"
	"errors"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *Application) GetActiveConfig(
	ctx context.Context, scope platformtx.TenantScope,
) (*ConfigDocument, error) {
	store, _, err := application.bind(ctx, scope, ScopeSnapshot, operationGet, permissionRead)
	if err != nil {
		return nil, err
	}
	document, err := store.GetActiveConfig(ctx)
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil || document.ID == uuid.Nil {
		return nil, ErrUnavailable
	}
	return &document, nil
}
