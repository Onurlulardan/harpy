package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func recoverExpiredJob(
	ctx context.Context,
	transaction pgx.Tx,
	authority JobAuthority,
	row expiredJobRow,
	decision recoveryDecision,
) error {
	result, err := transaction.Exec(
		ctx, recoverExpiredJobAttemptSQL,
		row.attemptID, row.jobID, row.leaseOwnerID, row.leaseGeneration,
		row.recoveredAt, decision.outcome, decision.errorClass,
		decision.errorCode, decision.errorDigest, decision.receiptID,
	)
	if err != nil {
		return fmt.Errorf("recover expired job attempt: %w", err)
	}
	if result.RowsAffected() != 1 {
		return ErrJobLeaseLost
	}
	arguments := append(
		authorityArguments(authority),
		row.jobID, row.leaseOwnerID, row.leaseGeneration,
		string(decision.state), decision.errorClass,
		decision.errorCode, decision.errorDigest, row.recoveredAt,
	)
	result, err = transaction.Exec(ctx, recoverExpiredJobSQL, arguments...)
	if err != nil {
		return fmt.Errorf("recover expired job: %w", err)
	}
	if result.RowsAffected() != 1 {
		return ErrJobLeaseLost
	}
	return nil
}
