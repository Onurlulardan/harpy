package ports

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

type TaskStore interface {
	InsertTask(context.Context, domain.Task) error
	LoadTaskForUpdate(context.Context, domain.ID) (domain.Task, error)
	PersistTask(context.Context, domain.Task, int64) error
	LoadLegacyTaskMetadataForUpdate(context.Context, domain.ID) (domain.LegacyTaskMetadata, error)
	UpsertLegacyTaskMetadata(context.Context, domain.ID, domain.LegacyTaskMetadata) error
	CountBoardTasksForUpdate(context.Context, domain.ID) (int, error)
	LoadRankedTasksForUpdate(context.Context, domain.ID) ([]domain.RankedTask, error)
	RebalanceTasks(context.Context, domain.ID, []domain.RankedTask) error
	LoadResourceLinksForUpdate(context.Context, domain.ID) ([]domain.ResourceLink, error)
	InsertResourceLink(context.Context, domain.ResourceLink) error
	ReplaceResourceLinks(context.Context, domain.ID, []domain.ResourceLink) error
	ReplaceAssignees(context.Context, domain.ID, []domain.Principal, int64) error
	InsertLabel(context.Context, domain.Label) error
	LoadLabelForUpdate(context.Context, domain.ID) (domain.Label, error)
	LoadLabelsForUpdate(context.Context, domain.ID) ([]domain.Label, error)
	PersistLabel(context.Context, domain.Label, int64) error
	PersistLabelArchive(context.Context, domain.ID, int64, *time.Time) error
	LoadTaskLabelIDsForUpdate(context.Context, domain.ID) ([]domain.ID, error)
	PersistTaskLabelLink(context.Context, domain.ID, domain.ID, bool) error
	InsertChecklistItem(context.Context, domain.ChecklistItem) error
	LoadChecklistForUpdate(context.Context, domain.ID) ([]domain.ChecklistItem, error)
	PersistChecklistItem(context.Context, domain.ChecklistItem, int64) error
	PersistChecklistArchive(context.Context, domain.ID, int64, *time.Time) error
}
