package schemaexecutor

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"time"
)

func (executor *Executor) recordFailure(ctx context.Context, lease Lease, cause error) error {
	kind, code := FaultDetails(cause)
	state := FailureTerminal
	if kind == FaultBlocked {
		state = FailureBlocked
	} else if kind == FaultTransient && lease.StepAttempt() < MaximumAttempts {
		state = FailureRetryable
	}
	detail := sha256.Sum256([]byte(code))
	request := FailureRequest{
		Lease: lease, State: state, Code: code,
		RedactedDetailSHA256: hex.EncodeToString(detail[:]),
	}
	if state == FailureRetryable {
		request.RetryDelay = retryDelay(lease)
	}
	return executor.persistence.RecordFailure(ctx, request)
}

func retryDelay(lease Lease) time.Duration {
	maximum := 5 * time.Second * time.Duration(1<<min(lease.StepAttempt()-1, 8))
	if maximum > 15*time.Minute {
		maximum = 15 * time.Minute
	}
	digest := sha256.Sum256([]byte(lease.LeaseID()))
	value := binary.BigEndian.Uint64(digest[:8])
	return time.Duration(value % uint64(maximum+1))
}
