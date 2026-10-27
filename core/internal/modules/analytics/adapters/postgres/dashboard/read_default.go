package dashboard

import (
	"context"
	_ "embed"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
)

//go:embed sql/tenant/get_default.sql
var getDefaultSQL string

func (storage *store) ReadDefault(
	ctx context.Context, surface string,
) (dashboardapp.Snapshot, error) {
	if !storage.usable(ctx) {
		return dashboardapp.Snapshot{}, dashboardapp.ErrUnavailable
	}
	if !surfacePattern.MatchString(surface) {
		return dashboardapp.Snapshot{}, dashboardapp.ErrInvalid
	}
	return scanSnapshot(
		storage.queryRow(ctx, getDefaultSQL, storage.actorID, surface),
		storage.notFound,
	)
}
