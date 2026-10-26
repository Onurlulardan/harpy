package postgres

import (
	"context"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func (repository *repository) ReleaseJobRetention(
	ctx context.Context,
	value jobsv1.RetentionRelease,
) (bool, error) {
	if repository == nil || !repository.authority.matchesRetentionRelease(value) {
		return false, ErrAuthorityMismatch
	}
	digest, err := value.SHA256()
	if err != nil {
		return false, fmt.Errorf("validate job retention release: %w", err)
	}
	jobID, err := uuid.Parse(value.JobID)
	if err != nil {
		return false, fmt.Errorf("parse retained job ID: %w", err)
	}
	retainUntil, err := time.Parse(time.RFC3339Nano, value.RetainUntil)
	if err != nil {
		return false, fmt.Errorf("parse job retention deadline: %w", err)
	}
	var released bool
	err = repository.database.QueryRow(
		ctx, releaseJobRetentionSQL, jobID, repository.authority.scopeID,
		repository.authority.groupID, repository.authority.groupGeneration,
		repository.authority.domain, repository.authority.generation,
		value.ReleaseRevision-1, digest, retainUntil,
	).Scan(&released)
	if err != nil {
		return false, fmt.Errorf("release async job retention: %w", err)
	}
	return released, nil
}
