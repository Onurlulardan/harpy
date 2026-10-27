package ports

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

type ProjectionItem struct {
	ID, SourceID, SourceVersionID                                     domain.ID
	SourceOwner                                                       domain.SourceOwner
	SourceKind, Surface, Summary                                      string
	VisibleTo                                                         []domain.Principal
	SourceRevision, ReferenceEpoch, PermissionRevision, EventSequence int64
	DueAt                                                             *time.Time
	Window                                                            *domain.CalendarWindow
	Visible                                                           bool
}

type ProjectionStore interface {
	LockCursor(context.Context, domain.SourceOwner, string, domain.ID) (domain.ProjectionCursor, error)
	PersistProjection(context.Context, ProjectionItem) error
	AdvanceCursor(context.Context, domain.SourceOwner, string, domain.ID, int64) error
	HideProjection(context.Context, domain.ID, int64) error
	FlagRepair(context.Context, domain.SourceOwner, string, domain.ID, int64) error
	ReplaceSourceProjection(context.Context, domain.SourceOwner, string, domain.ID, []ProjectionItem, int64) error
}
