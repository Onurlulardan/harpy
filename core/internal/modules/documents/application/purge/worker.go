package purge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Worker struct {
	binder  StoreBinder
	jobs    Jobs
	command *Command
}

func NewWorker(binder StoreBinder, jobs Jobs, ids foundation.IDSource,
	clock foundation.Clock,
) (*Worker, error) {
	for _, dependency := range []any{binder, jobs, ids, clock} {
		if foundation.NilPort(dependency) {
			return nil, fmt.Errorf("documents purge dependencies are invalid")
		}
	}
	command, err := NewCommand(ids, clock)
	if err != nil {
		return nil, err
	}
	return &Worker{binder: binder, jobs: jobs, command: command}, nil
}

func (worker *Worker) Run(ctx context.Context, runner JobRunner,
	tenantID, itemID domain.ID,
) (Result, error) {
	if worker == nil || foundation.NilPort(runner) || !tenantID.Valid() ||
		!itemID.UUIDv7() || !platformtx.NetworkAllowed(ctx) {
		return Result{}, ErrInvalid
	}
	var result Result
	err := runner.InRepeatableReadJobTenant(ctx, tenantID,
		func(scoped context.Context, scope platformtx.JobTenantScope) error {
			store, bindErr := worker.binder.BindJob(scoped, scope)
			if bindErr != nil || store == nil {
				return ErrUnavailable
			}
			execution, bindErr := worker.command.Execute(scoped, store, itemID)
			if bindErr != nil {
				return bindErr
			}
			result = execution.Result
			for _, deletion := range execution.Deletions {
				payload, _ := json.Marshal(struct {
					DeletionID string `json:"deletion_id"`
				}{DeletionID: deletion.DeletionID.String()})
				if bindErr = worker.jobs.EnqueueDeletion(scoped, scope, deletion.JobID,
					deletion.DeletionID, payload); bindErr != nil {
					return bindErr
				}
			}
			return nil
		})
	return result, mapError(err)
}
