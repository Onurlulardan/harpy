package systemtransfer

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransferapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

type WorkerTerminal struct {
	State                                                 string
	Processed, Created, Updated, Skipped, Rejected, Bytes int64
}

func (store *WorkerStore) CommitTerminal(ctx context.Context,
	scope tenantScope, evidence publicresource.ActorEvidence,
	attempt systemtransferapp.Attempt, success bool,
) (WorkerTerminal, error) {
	if attempt.Validate() != nil || scope.TenantID() != attempt.TenantID ||
		scope.ActorID() != attempt.ActorID {
		return WorkerTerminal{}, systemtransferapp.ErrInvalidAttempt
	}
	token, authorization := int64(0), attempt.AuthorizationVersion
	var err error
	if success {
		token, authorization, err = workerEvidence(scope, evidence, attempt)
		if err != nil {
			return WorkerTerminal{}, err
		}
	}
	bound, err := store.bind(ctx, scope)
	if err != nil {
		return WorkerTerminal{}, err
	}
	schema, err := bound.database.CurrentSchema(ctx)
	if err != nil {
		return WorkerTerminal{}, systemtransferapp.ErrWorkerUnavailable
	}
	var result WorkerTerminal
	err = bound.database.QueryRow(ctx, finishImportWorkerSQL,
		schema, attempt.TenantID, attempt.ActorID, attempt.ID, attempt.JobID,
		attempt.AttemptID, attempt.LeaseOwnerID, attempt.LeaseGeneration,
		string(attempt.Resource), token, authorization, success,
	).Scan(&result.State, &result.Processed, &result.Created, &result.Updated,
		&result.Skipped, &result.Rejected, &result.Bytes)
	if err != nil {
		return WorkerTerminal{}, mapWorkerError(err)
	}
	want := "failed"
	if success {
		want = "completed"
	}
	if result.State != want {
		return WorkerTerminal{}, systemtransferapp.ErrWorkerState
	}
	return result, nil
}
