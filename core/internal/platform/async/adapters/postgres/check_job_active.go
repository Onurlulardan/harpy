package postgres

import (
	"context"
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func (repository *workerRepository) CheckJobActive(ctx context.Context,
	jobID uuid.UUID, authority jobsv1.Authority,
) error {
	if repository == nil || !repository.valid() || ctx == nil || ctx.Err() != nil ||
		jobID == uuid.Nil || authority.Validate() != nil ||
		authority.ScopeKind != jobsv1.ScopeTenant ||
		authority.WriterOrigin != jobsv1.WriterV2 {
		return ErrJobNotControllable
	}
	scopeID, err := uuid.Parse(authority.ScopeID)
	if err != nil {
		return ErrJobNotControllable
	}
	rows, err := repository.database.Query(ctx, checkJobActiveSQL, jobID, scopeID,
		authority.GroupID, authority.GroupGeneration, authority.Domain,
		authority.CutoverGeneration)
	if err != nil {
		return fmt.Errorf("check active job: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return fmt.Errorf("check active job: %w", err)
		}
		return ErrJobNotControllable
	}
	var found uuid.UUID
	if err = rows.Scan(&found); err != nil || found != jobID || rows.Next() {
		return ErrJobNotControllable
	}
	return rows.Err()
}
