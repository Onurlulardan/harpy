package domain

import (
	"strings"
	"time"
)

type Task struct {
	ID, BoardID, ColumnID ID
	CreatorID             ID
	Kind                  TaskKind
	Status                TaskStatus
	Priority              TaskPriority
	Title, Description    string
	Position, RowVersion  int64
	StartsAt              *time.Time
	DueAt                 *time.Time
	DeletedAt             *time.Time
}

func (task Task) Valid() bool {
	return task.ID.Valid() && task.BoardID.Valid() && task.ColumnID.Valid() &&
		task.CreatorID.Valid() &&
		task.Kind.Valid() && task.Status.Valid() && task.Priority.Valid() &&
		strings.TrimSpace(task.Title) != "" &&
		len(task.Title) <= 300 && len(task.Description) <= 16_000 &&
		task.Position > 0 && task.RowVersion > 0
}

type Label struct {
	ID, BoardID ID
	Name, Color string
	RowVersion  int64
	DeletedAt   *time.Time
}

func (label Label) Valid() bool {
	return label.ID.Valid() && label.BoardID.Valid() && strings.TrimSpace(label.Name) != "" &&
		len(label.Name) <= 100 && len(label.Color) == 7 && label.Color[0] == '#' && label.RowVersion > 0
}

type ChecklistItem struct {
	ID, TaskID ID
	Text       string
	Completed  bool
	Position   int64
	RowVersion int64
	DeletedAt  *time.Time
}

func (item ChecklistItem) Valid() bool {
	return item.ID.Valid() && item.TaskID.Valid() && strings.TrimSpace(item.Text) != "" &&
		len(item.Text) <= 1000 && item.Position > 0 && item.RowVersion > 0
}
