package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

//go:embed sql/tenant/rebalance_task.sql
var rebalanceTaskSQL string

func (store *Store) RebalanceTasks(ctx context.Context, columnID domain.ID, tasks []domain.RankedTask) error {
	if !columnID.Valid() || len(tasks) > domain.MaxColumnTasks {
		return application.ErrInvalid
	}
	for _, task := range tasks {
		if !task.ID.Valid() || task.Position < 1 {
			return application.ErrInvalid
		}
		if err := executeOne(ctx, store, rebalanceTaskSQL, task.ID, columnID, task.Position); err != nil {
			return err
		}
	}
	return nil
}
