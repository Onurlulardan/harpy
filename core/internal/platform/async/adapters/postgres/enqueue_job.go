package postgres

import (
	"context"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func (repository *repository) EnqueueJob(
	ctx context.Context,
	value jobsv1.Enqueue,
) (EnqueueResult, error) {
	if repository == nil || !repository.authority.matchesJob(value) {
		return EnqueueResult{}, ErrAuthorityMismatch
	}
	if err := value.Validate(repository.jobs); err != nil {
		return EnqueueResult{}, fmt.Errorf("validate job enqueue: %w", err)
	}
	policy, _ := repository.jobs.Resolve(value.Kind, value.PayloadRevision)
	jobID, runAt, retryID, err := jobEnqueueValues(value)
	if err != nil {
		return EnqueueResult{}, err
	}
	effectClass := string(policy.EffectClass)
	result, err := repository.database.Execute(
		ctx, enqueueJobSQL, jobID, repository.authority.scopeID, "tenant",
		repository.authority.scopeID, repository.authority.groupID,
		repository.authority.groupGeneration, repository.authority.domain,
		repository.authority.generation, "v2", value.Kind, value.PayloadRevision,
		value.PayloadReference, value.PayloadSHA256, value.DedupeKey, effectClass,
		value.EffectIdempotencyKey, value.Priority, runAt, value.MaximumAttempts, retryID,
	)
	if err != nil {
		return EnqueueResult{}, fmt.Errorf("insert async job: %w", err)
	}
	if result.RowsAffected() == 1 {
		return EnqueueResult{ID: jobID.String(), Created: true}, nil
	}
	return repository.findJobEnqueue(ctx, value, jobID, effectClass, retryID, runAt)
}

func jobEnqueueValues(value jobsv1.Enqueue) (uuid.UUID, time.Time, *uuid.UUID, error) {
	jobID, err := uuid.Parse(value.JobID)
	if err != nil {
		return uuid.Nil, time.Time{}, nil, fmt.Errorf("parse job ID: %w", err)
	}
	runAt, err := time.Parse(time.RFC3339Nano, value.RunAt)
	if err != nil {
		return uuid.Nil, time.Time{}, nil, fmt.Errorf("parse job run time: %w", err)
	}
	var retryID *uuid.UUID
	if value.RetryOfJobID != nil {
		parsed, parseErr := uuid.Parse(*value.RetryOfJobID)
		if parseErr != nil {
			return uuid.Nil, time.Time{}, nil, fmt.Errorf("parse retry job ID: %w", parseErr)
		}
		retryID = &parsed
	}
	return jobID, runAt, retryID, nil
}

func (repository *repository) findJobEnqueue(
	ctx context.Context,
	value jobsv1.Enqueue,
	jobID uuid.UUID,
	effectClass string,
	retryID *uuid.UUID,
	runAt time.Time,
) (EnqueueResult, error) {
	var existingID string
	var exact bool
	err := repository.database.QueryRow(
		ctx, findJobEnqueueSQL, jobID, repository.authority.scopeID,
		repository.authority.groupID, repository.authority.domain, value.Kind,
		value.DedupeKey, repository.authority.groupGeneration,
		repository.authority.generation, value.PayloadRevision, value.PayloadSHA256,
		effectClass, value.EffectIdempotencyKey, value.Priority,
		value.MaximumAttempts, retryID, runAt, value.PayloadReference,
	).Scan(&existingID, &exact)
	if err != nil {
		return EnqueueResult{}, fmt.Errorf("read existing async job: %w", err)
	}
	if !exact {
		return EnqueueResult{}, ErrIdempotencyConflict
	}
	return EnqueueResult{ID: existingID}, nil
}
