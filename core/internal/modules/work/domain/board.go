package domain

import (
	"strings"
	"time"
)

type Lifecycle string

const (
	LifecycleActive   Lifecycle = "ACTIVE"
	LifecycleArchived Lifecycle = "ARCHIVED"
)

type Board struct {
	ID, CreatorID     ID
	Name, Description string
	Lifecycle         Lifecycle
	RowVersion        int64
	CreatedAt         time.Time
}

func (board Board) Valid() bool {
	return board.ID.Valid() && board.CreatorID.Valid() && strings.TrimSpace(board.Name) != "" &&
		len(board.Name) <= 200 && len(board.Description) <= 16_000 &&
		(board.Lifecycle == LifecycleActive ||
			board.Lifecycle == LifecycleArchived) && board.RowVersion > 0 && !board.CreatedAt.IsZero()
}

type Column struct {
	ID, BoardID ID
	Name, Color string
	Status      TaskStatus
	Position    int64
	RowVersion  int64
}

func (column Column) Valid() bool {
	return column.ID.Valid() && column.BoardID.Valid() && strings.TrimSpace(column.Name) != "" &&
		len(column.Name) <= 200 && ValidColor(column.Color) &&
		column.Status.Valid() && column.Status != StatusCancelled &&
		column.Position > 0 && column.RowVersion > 0
}
