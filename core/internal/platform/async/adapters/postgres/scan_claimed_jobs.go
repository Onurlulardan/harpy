package postgres

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func scanClaimedJobRows(rows pgx.Rows, limit int) ([]jobClaimRow, error) {
	defer rows.Close()
	claimed := make([]jobClaimRow, 0, limit)
	for rows.Next() {
		var row jobClaimRow
		if err := rows.Scan(
			&row.jobID,
			&row.tenantID,
			&row.scopeKind,
			&row.scopeID,
			&row.groupID,
			&row.groupGeneration,
			&row.domain,
			&row.generation,
			&row.kind,
			&row.payloadRevision,
			&row.payloadReference,
			&row.payloadSHA256,
			&row.effectClass,
			&row.effectKey,
			&row.priority,
			&row.runAt,
			&row.attempt,
			&row.maximumAttempts,
			&row.leaseOwnerID,
			&row.leaseGeneration,
			&row.startedAt,
			&row.leaseUntil,
		); err != nil {
			return nil, fmt.Errorf("scan claimed job: %w", err)
		}
		claimed = append(claimed, row)
		if len(claimed) > limit {
			return nil, ErrJobClaimInvalid
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate claimed jobs: %w", err)
	}
	return claimed, nil
}
