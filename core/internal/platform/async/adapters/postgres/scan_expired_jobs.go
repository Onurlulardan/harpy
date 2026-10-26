package postgres

import (
	"fmt"

	"github.com/jackc/pgx/v5"
)

func scanExpiredJobs(rows pgx.Rows, limit int) ([]expiredJobRow, error) {
	defer rows.Close()
	result := make([]expiredJobRow, 0, limit)
	for rows.Next() {
		var row expiredJobRow
		if err := rows.Scan(
			&row.jobID, &row.attemptID, &row.leaseOwnerID,
			&row.leaseGeneration, &row.attempt, &row.maximumAttempts,
			&row.effectClass, &row.cancelRequested, &row.effectStarted,
			&row.effectID, &row.requestSHA256, &row.receiptCount,
			&row.receiptID, &row.receiptStatus, &row.receiptErrorClass,
			&row.receiptCode, &row.recoveredAt,
		); err != nil {
			return nil, fmt.Errorf("scan expired job: %w", err)
		}
		result = append(result, row)
		if len(result) > limit {
			return nil, ErrJobRecoveryLimit
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expired jobs: %w", err)
	}
	return result, nil
}
