package discounteligibility

import (
	"context"
	"fmt"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	"github.com/google/uuid"
)

type authorities struct {
	tenant, global uuid.UUID
}

func (storage *store) authorityIDs(ctx context.Context) (authorities, error) {
	var result authorities
	err := storage.queryRow(ctx, resolveAuthoritySQL, storage.tenantID).Scan(
		&result.tenant, &result.global,
	)
	if err != nil {
		if storage.notFound(err) {
			return authorities{}, discountapp.ErrQueryDependency
		}
		return authorities{}, fmt.Errorf(
			"%w: resolve discount authorities", discountapp.ErrQueryStorage,
		)
	}
	if result.tenant == uuid.Nil || result.global == uuid.Nil {
		return authorities{}, discountapp.ErrQueryDependency
	}
	return result, nil
}
