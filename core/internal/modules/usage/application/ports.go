package application

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

type RecordRequest struct {
	EnvironmentID, TenantID, ActorID uuid.UUID
	SourceIdentity                   [32]byte
	Method, RoutePattern             string
	StatusCode                       int
	DurationMicros                   int64
	OccurredAt                       time.Time
}

type RecordResult struct {
	FactID   uuid.UUID
	Replayed bool
}

type Store interface {
	Record(context.Context, domain.Observation, [32]byte) (RecordResult, error)
	Overview(context.Context, Access, domain.TimeRange, time.Time) (domain.Overview, error)
	ActiveUsers(context.Context, Access, time.Time) (domain.ActiveUsers, error)
	TopUsers(context.Context, Access, domain.RankedQuery) ([]domain.TopUser, error)
	UserSummary(context.Context, Access, string, domain.TimeRange) (domain.UserSummary, error)
	UserTimeline(context.Context, Access, domain.TimelineQuery) (domain.TimelinePage, error)
	Features(context.Context, Access, domain.RankedQuery) ([]domain.FeatureUsage, error)
	Heatmap(context.Context, Access, domain.TimeRange) ([]domain.HeatmapCell, error)
	Trend(context.Context, Access, domain.TimeRange, domain.Granularity) ([]domain.TrendPoint, error)
}

type Clock interface{ Now() time.Time }
type IDGenerator interface{ NewV7() (uuid.UUID, error) }
