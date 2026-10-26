package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func insertJobAttempts(
	ctx context.Context,
	transaction pgx.Tx,
	claimed []jobClaimRow,
) ([]jobClaimRow, error) {
	if len(claimed) == 0 {
		return claimed, nil
	}
	jobIDs := make([]uuid.UUID, len(claimed))
	scopeKinds := make([]string, len(claimed))
	scopeIDs := make([]uuid.UUID, len(claimed))
	attempts := make([]int32, len(claimed))
	leaseGenerations := make([]int64, len(claimed))
	leaseOwners := make([]uuid.UUID, len(claimed))
	startedAt := make([]time.Time, len(claimed))
	positions := make(map[uuid.UUID]int, len(claimed))
	for index := range claimed {
		row := claimed[index]
		if _, duplicate := positions[row.jobID]; duplicate {
			return nil, ErrJobClaimInvalid
		}
		positions[row.jobID] = index
		jobIDs[index] = row.jobID
		scopeKinds[index] = row.scopeKind
		scopeIDs[index] = row.scopeID
		attempts[index] = row.attempt
		leaseGenerations[index] = row.leaseGeneration
		leaseOwners[index] = row.leaseOwnerID
		startedAt[index] = row.startedAt
	}
	rows, err := transaction.Query(
		ctx,
		insertJobAttemptsSQL,
		jobIDs,
		scopeKinds,
		scopeIDs,
		attempts,
		leaseGenerations,
		leaseOwners,
		startedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert job attempts: %w", err)
	}
	defer rows.Close()
	inserted := make(map[uuid.UUID]struct{}, len(claimed))
	for rows.Next() {
		var attemptID, jobID uuid.UUID
		if err := rows.Scan(&attemptID, &jobID); err != nil {
			return nil, fmt.Errorf("scan inserted job attempt: %w", err)
		}
		position, expected := positions[jobID]
		if !expected || attemptID == uuid.Nil {
			return nil, ErrJobClaimInvalid
		}
		if _, duplicate := inserted[jobID]; duplicate {
			return nil, ErrJobClaimInvalid
		}
		inserted[jobID] = struct{}{}
		claimed[position].attemptID = attemptID
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inserted job attempts: %w", err)
	}
	if len(inserted) != len(claimed) {
		return nil, ErrJobClaimInvalid
	}
	return claimed, nil
}
