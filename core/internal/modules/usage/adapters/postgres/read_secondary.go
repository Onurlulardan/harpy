package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func (store *Store) UserTimeline(ctx context.Context, access application.Access,
	query domain.TimelineQuery,
) (domain.TimelinePage, error) {
	if !allows(access, application.OperationTenantTimeline, application.OperationPlatformTimeline) {
		return domain.TimelinePage{}, domain.ErrForbidden
	}
	document, err := store.read(ctx, userTimelineSQL,
		withArguments(access, query.Subject, query.Page, query.Limit)...)
	if err != nil {
		return domain.TimelinePage{}, err
	}
	return decodeTimeline(document)
}

func (store *Store) Features(ctx context.Context, access application.Access,
	query domain.RankedQuery,
) ([]domain.FeatureUsage, error) {
	if !allows(access, application.OperationTenantFeatures, application.OperationPlatformFeatures) {
		return nil, domain.ErrForbidden
	}
	document, err := store.read(ctx, featuresSQL,
		withArguments(access, query.Range.Start, query.Range.End, query.Limit)...)
	if err != nil {
		return nil, err
	}
	return decodeFeatures(document)
}

func (store *Store) Heatmap(ctx context.Context, access application.Access,
	window domain.TimeRange,
) ([]domain.HeatmapCell, error) {
	if !allows(access, application.OperationTenantHeatmap, application.OperationPlatformHeatmap) {
		return nil, domain.ErrForbidden
	}
	document, err := store.read(ctx, heatmapSQL,
		withArguments(access, window.Start, window.End)...)
	if err != nil {
		return nil, err
	}
	return decodeHeatmap(document)
}

func (store *Store) Trend(ctx context.Context, access application.Access,
	window domain.TimeRange, granularity domain.Granularity,
) ([]domain.TrendPoint, error) {
	if !allows(access, application.OperationTenantTrend, application.OperationPlatformTrend) {
		return nil, domain.ErrForbidden
	}
	if !domain.ValidGranularity(granularity) {
		return nil, domain.ErrInvalid
	}
	document, err := store.read(ctx, trendSQL,
		withArguments(access, window.Start, window.End, string(granularity))...)
	if err != nil {
		return nil, err
	}
	return decodeTrend(document)
}

func (store *Store) read(ctx context.Context, statement string, arguments ...any) ([]byte, error) {
	var document []byte
	err := store.within(ctx, SnapshotTransaction, func(database Database) error {
		return database.QueryRow(ctx, statement, arguments...).Scan(&document)
	})
	return document, store.normalize(err)
}
