package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (repository *workerRepository) ResolveStaleJobAuthority(
	ctx context.Context,
	jobID uuid.UUID,
) (JobAuthority, error) {
	if ctx == nil {
		return JobAuthority{}, ErrWorkerContextInvalid
	}
	if !repository.valid() {
		return JobAuthority{}, ErrWorkerStoreUnavailable
	}
	if jobID == uuid.Nil {
		return JobAuthority{}, ErrStaleJobDispositionInvalid
	}
	rows, err := repository.database.Query(ctx, resolveStaleJobAuthoritySQL, jobID)
	if err != nil {
		return JobAuthority{}, fmt.Errorf("resolve stale job authority: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return JobAuthority{}, fmt.Errorf("resolve stale job authority: %w", err)
		}
		return JobAuthority{}, ErrStaleJobDispositionConflict
	}
	var scopeKind, groupID, domain string
	var scopeID uuid.UUID
	var groupGeneration, domainGeneration int64
	if err := rows.Scan(
		&scopeKind, &scopeID, &groupID, &groupGeneration,
		&domain, &domainGeneration,
	); err != nil {
		return JobAuthority{}, fmt.Errorf("scan stale job authority: %w", err)
	}
	if rows.Next() {
		return JobAuthority{}, ErrStaleJobDispositionConflict
	}
	if err := rows.Err(); err != nil {
		return JobAuthority{}, fmt.Errorf("iterate stale job authority: %w", err)
	}
	authority, err := newJobAuthority(
		repository.identity, scopeKind, scopeID, groupID,
		groupGeneration, domain, domainGeneration,
	)
	if err != nil {
		return JobAuthority{}, fmt.Errorf("decode stale job authority: %w", err)
	}
	return authority, nil
}
