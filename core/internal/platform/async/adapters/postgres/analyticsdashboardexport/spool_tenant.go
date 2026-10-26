package analyticsdashboardexport

import (
	"context"
	"fmt"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (store *Store) AnalyticsDashboardExportSpoolTenant(
	ctx context.Context, exportID uuid.UUID,
) (uuid.UUID, bool, error) {
	if store == nil || store.database == nil || ctx == nil || ctx.Err() != nil {
		return uuid.Nil, false,
			platformasync.ErrAnalyticsDashboardExportSourceUnavailable
	}
	if exportID == uuid.Nil || exportID.Version() != uuid.Version(7) ||
		exportID.Variant() != uuid.RFC4122 {
		return uuid.Nil, false, platformasync.ErrAnalyticsDashboardExportEvidenceInvalid
	}
	var stored pgtype.UUID
	if err := store.database.QueryRow(ctx, spoolTenantSQL, exportID).Scan(&stored); err != nil {
		return uuid.Nil, false,
			fmt.Errorf("resolve analytics dashboard export spool tenant: %w", err)
	}
	if !stored.Valid {
		return uuid.Nil, false, nil
	}
	tenantID := uuid.UUID(stored.Bytes)
	if tenantID == uuid.Nil {
		return uuid.Nil, false, platformasync.ErrAnalyticsDashboardExportEvidenceInvalid
	}
	return tenantID, true, nil
}
