package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/grouprehearsal"
	"github.com/jackc/pgx/v5"
)

type GroupRehearsalDatabase interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type groupRehearsalRecorder struct {
	database GroupRehearsalDatabase
}

func NewGroupRehearsalRecorder(
	database GroupRehearsalDatabase,
) (grouprehearsal.Recorder, error) {
	if database == nil {
		return nil, fmt.Errorf("billing Forms rehearsal database is required")
	}
	return &groupRehearsalRecorder{database: database}, nil
}
