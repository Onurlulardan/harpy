package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func schemaExecutorInsertAttempt(
	ctx context.Context,
	transaction pgx.Tx,
	lease schemaExecutorLeaseView,
	state string,
	errorCode *string,
	redactedDetail *string,
) (time.Time, error) {
	outcome := schemaExecutorOutcomeHash(lease, state, errorCode)
	var finishedAt time.Time
	err := transaction.QueryRow(ctx, schemaExecutorInsertAttemptSQL,
		lease.JobID(), lease.StepOrdinal(), lease.StepAttempt(),
		lease.PlanSHA256(), lease.StepSHA256(), lease.LeaseID(), lease.ExecutorID(),
		state, lease.StepStartedAt(), outcome, errorCode, redactedDetail,
	).Scan(&finishedAt)
	return finishedAt, err
}

func schemaExecutorOutcomeHash(
	lease schemaExecutorLeaseView,
	state string,
	errorCode *string,
) string {
	code := ""
	if errorCode != nil {
		code = *errorCode
	}
	parts := []string{
		lease.JobID(), lease.PlanSHA256(), lease.StepSHA256(),
		lease.LeaseID(), lease.ExecutorID(),
		strconv.Itoa(lease.JobAttempt()), strconv.Itoa(lease.StepAttempt()),
		state, code,
	}
	digest := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(digest[:])
}
