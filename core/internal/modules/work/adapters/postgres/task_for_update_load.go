package postgres

import (
	"context"
	_ "embed"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/get_task_for_update.sql
var getTaskForUpdateSQL string

func (store *Store) LoadTaskForUpdate(ctx context.Context, id domain.ID) (domain.Task, error) {
	if !id.Valid() {
		return domain.Task{}, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return domain.Task{}, err
	}
	var taskID, boardID, columnID, creatorID uuid.UUID
	var kind, status, priority, title, description string
	var position, version int64
	var starts, due, deleted *time.Time
	err := store.database.QueryRow(ctx, getTaskForUpdateSQL, id).Scan(&taskID, &boardID, &columnID,
		&creatorID, &kind, &status, &priority, &title, &description,
		&position, &version, &starts, &due, &deleted)
	if err != nil {
		return domain.Task{}, err
	}
	task := domain.Task{ID: domain.ID(taskID), BoardID: domain.ID(boardID), ColumnID: domain.ID(columnID),
		CreatorID: domain.ID(creatorID), Kind: domain.TaskKind(kind),
		Status: domain.TaskStatus(status), Title: title,
		Priority: domain.TaskPriority(priority), Description: description,
		Position: position, RowVersion: version}
	if starts != nil {
		value := starts.UTC()
		task.StartsAt = &value
	}
	if due != nil {
		value := due.UTC()
		task.DueAt = &value
	}
	if deleted != nil {
		value := deleted.UTC()
		task.DeletedAt = &value
	}
	if !task.Valid() {
		return domain.Task{}, application.ErrUnavailable
	}
	return task, nil
}
