package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/fullrollout"
	"github.com/jackc/pgx/v5"
)

type FullRolloutDatabase interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type fullRolloutRecorder struct{ database FullRolloutDatabase }

func NewFullRolloutRecorder(database FullRolloutDatabase) (fullrollout.Recorder, error) {
	if database == nil {
		return nil, fmt.Errorf("full rollout database is required")
	}
	return &fullRolloutRecorder{database: database}, nil
}
