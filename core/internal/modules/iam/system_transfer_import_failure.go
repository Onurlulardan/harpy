package iam

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

func (terminal *systemTransferImportJobTerminalRuntime) Fail(ctx context.Context,
	claim asyncpostgres.JobClaim,
	attempt systemtransfer.Attempt,
	failure systemtransfer.WorkerFailure,
) error {
	class, err := systemTransferImportErrorClass(failure.Class)
	if terminal == nil || err != nil || attempt.Validate() != nil ||
		failure.ActorID != attempt.ActorID || failure.Digest == [32]byte{} {
		return fmt.Errorf("finalize IAM system import failure: invalid input")
	}
	delay := time.Duration(0)
	if class == jobsv1.ErrorTransient && claim.Attempt() < claim.MaximumAttempts() {
		delay, err = platformasync.RetryDelay(claim.Attempt(), terminal.jitter)
		if err != nil {
			return fmt.Errorf("calculate IAM system import retry: %w", err)
		}
	}
	jobFailure, err := asyncpostgres.NewJobFailure(
		class, failure.Code, hex.EncodeToString(failure.Digest[:]), delay,
	)
	if err != nil {
		return fmt.Errorf("build IAM system import failure: %w", err)
	}
	return terminal.scopes.InFreshImportBatch(ctx, attempt,
		func(scoped context.Context, batch systemtransfer.BatchScope) error {
			writer, err := terminal.writer(batch.IAMScope())
			if err != nil {
				return err
			}
			result, err := writer.FailJobBeforeEffect(scoped, claim, jobFailure)
			if err != nil {
				return err
			}
			if result.State() != jobsv1.StateDead &&
				result.State() != jobsv1.StateCancelled {
				return nil
			}
			_, err = terminal.finisher.CommitTerminal(scoped, batch.IAMScope(),
				publicresource.ActorEvidence{}, attempt, false)
			return err
		})
}

func systemTransferImportErrorClass(class systemtransfer.WorkerFailureClass) (
	jobsv1.ErrorClass, error,
) {
	switch class {
	case systemtransfer.WorkerFailurePermanent:
		return jobsv1.ErrorPermanent, nil
	case systemtransfer.WorkerFailureTransient:
		return jobsv1.ErrorTransient, nil
	case systemtransfer.WorkerFailureCancelled:
		return jobsv1.ErrorCancelled, nil
	default:
		return "", fmt.Errorf("unknown IAM system import failure class")
	}
}

type systemTransferImportJitter struct{}

func (systemTransferImportJitter) Uint64N(bound uint64) uint64 {
	if bound == 0 {
		panic("IAM system import retry jitter bound is zero")
	}
	value, err := rand.Int(rand.Reader, new(big.Int).SetUint64(bound))
	if err != nil {
		panic("IAM system import retry entropy is unavailable")
	}
	return value.Uint64()
}
