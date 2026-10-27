package legalhold

import (
	"context"
	"fmt"

	legalholdapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/legalhold"
	"github.com/google/uuid"
)

func (storage *store) authorityID(ctx context.Context) (uuid.UUID, error) {
	if !storage.usable(ctx) {
		return uuid.Nil, legalholdapp.ErrConfiguration
	}
	var authorityID uuid.UUID
	if err := storage.database.QueryRow(
		ctx, resolveAuthoritySQL, storage.tenantID,
	).Scan(&authorityID); err != nil || authorityID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("resolve billing legal hold authority: %w", err)
	}
	return authorityID, nil
}
