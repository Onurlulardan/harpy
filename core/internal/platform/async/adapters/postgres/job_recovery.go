package postgres

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type expiredJobRow struct {
	jobID             uuid.UUID
	attemptID         uuid.UUID
	leaseOwnerID      uuid.UUID
	leaseGeneration   int64
	attempt           int32
	maximumAttempts   int32
	effectClass       string
	cancelRequested   bool
	effectStarted     pgtype.Timestamptz
	effectID          pgtype.UUID
	requestSHA256     pgtype.Text
	receiptCount      int64
	receiptID         pgtype.UUID
	receiptStatus     pgtype.Text
	receiptErrorClass pgtype.Text
	receiptCode       pgtype.Text
	recoveredAt       time.Time
}

type recoveryDecision struct {
	state       jobsv1.State
	outcome     string
	errorClass  any
	errorCode   any
	errorDigest any
	receiptID   any
}

type JobRecoveryResult struct {
	requeued  int
	succeeded int
	cancelled int
	dead      int
}

func (result JobRecoveryResult) Requeued() int  { return result.requeued }
func (result JobRecoveryResult) Succeeded() int { return result.succeeded }
func (result JobRecoveryResult) Cancelled() int { return result.cancelled }
func (result JobRecoveryResult) Dead() int      { return result.dead }
func (result JobRecoveryResult) Total() int {
	return result.requeued + result.succeeded + result.cancelled + result.dead
}

func recoveryError(class string, code string) recoveryDecision {
	digest := sha256.Sum256([]byte(code))
	return recoveryDecision{
		errorClass: class, errorCode: code,
		errorDigest: hex.EncodeToString(digest[:]),
	}
}

func decideExpiredJob(row expiredJobRow) recoveryDecision {
	if row.effectStarted.Valid {
		return decideExpiredExternalJob(row)
	}
	if row.cancelRequested {
		decision := recoveryError("cancelled", "JOB_CANCELLED_LEASE_EXPIRED")
		decision.state, decision.outcome = jobsv1.StateCancelled, "lease_expired"
		return decision
	}
	if row.attempt < row.maximumAttempts {
		decision := recoveryError("transient", "JOB_LEASE_EXPIRED")
		decision.state, decision.outcome = jobsv1.StateQueued, "lease_expired"
		return decision
	}
	decision := recoveryError("transient", "JOB_ATTEMPTS_EXHAUSTED")
	decision.state, decision.outcome = jobsv1.StateDead, "dead"
	return decision
}

func decideExpiredExternalJob(row expiredJobRow) recoveryDecision {
	if row.receiptCount != 1 || !row.receiptID.Valid {
		return recoveryDecision{
			state: jobsv1.StateDead, outcome: "dead",
			errorCode: jobsv1.UnknownOutcomeCode,
		}
	}
	receiptID := uuid.UUID(row.receiptID.Bytes)
	if row.receiptStatus.String == "succeeded" {
		return recoveryDecision{
			state: jobsv1.StateSucceeded, outcome: "succeeded", receiptID: receiptID,
		}
	}
	decision := recoveryError(row.receiptErrorClass.String, row.receiptCode.String)
	decision.receiptID = receiptID
	switch row.receiptErrorClass.String {
	case "cancelled":
		decision.state, decision.outcome = jobsv1.StateCancelled, "lease_expired"
	case "transient", "conflict":
		if row.attempt < row.maximumAttempts {
			decision.state, decision.outcome = jobsv1.StateQueued, "lease_expired"
		} else {
			decision.state, decision.outcome = jobsv1.StateDead, "dead"
		}
	default:
		decision.state, decision.outcome = jobsv1.StateDead, "dead"
	}
	return decision
}
