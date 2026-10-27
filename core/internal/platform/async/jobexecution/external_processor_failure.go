package jobexecution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

func (processor *ExternalProcessor) failBeforeEffect(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
	code string,
	digest string,
) error {
	failure, err := asyncpostgres.NewJobFailure(
		jobsv1.ErrorPermanent, code, digest, 0,
	)
	if err != nil {
		return fmt.Errorf("build pre-effect failure: %w", err)
	}
	return processor.persistPreEffectFailure(ctx, claim, failure)
}

func (processor *ExternalProcessor) cancelBeforeEffect(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
) error {
	failure, err := asyncpostgres.NewJobFailure(
		jobsv1.ErrorCancelled, codeJobCancellationRequested,
		stableFailureDigest(codeJobCancellationRequested), 0,
	)
	if err != nil {
		return fmt.Errorf("build cancellation failure: %w", err)
	}
	return processor.persistPreEffectFailure(ctx, claim, failure)
}

func (processor *ExternalProcessor) persistPreEffectFailure(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
	failure asyncpostgres.JobFailure,
) error {
	finalizeContext, cancel := boundedFinalizationContext(ctx)
	defer cancel()
	if _, err := processor.preEffect.FailBeforeEffect(
		finalizeContext, claim, failure,
	); err != nil {
		return fmt.Errorf("persist pre-effect job failure: %w", err)
	}
	return nil
}

func (processor *ExternalProcessor) failureFromReceipt(
	receipt effectsv1.Receipt,
	attempt int,
) (asyncpostgres.JobFailure, error) {
	if receipt.ErrorClass == nil {
		return asyncpostgres.JobFailure{}, fmt.Errorf("failed receipt has no error class")
	}
	errorClass := jobsv1.ErrorClass(*receipt.ErrorClass)
	delay := platformasyncRetryDelay(processor, errorClass, attempt)
	if delay.err != nil {
		return asyncpostgres.JobFailure{}, delay.err
	}
	document, err := receipt.MarshalJSON()
	if err != nil {
		return asyncpostgres.JobFailure{}, err
	}
	digest, err := canonicaljsonv1.DigestHex(document, effectsv1.MaximumDocumentBytes)
	if err != nil {
		return asyncpostgres.JobFailure{}, err
	}
	return asyncpostgres.NewJobFailure(errorClass, receipt.Code, digest, delay.value)
}

type retryDelayResult struct {
	value time.Duration
	err   error
}

func platformasyncRetryDelay(
	processor *ExternalProcessor,
	errorClass jobsv1.ErrorClass,
	attempt int,
) retryDelayResult {
	if errorClass != jobsv1.ErrorTransient && errorClass != jobsv1.ErrorConflict ||
		attempt >= processor.policy.MaximumAttempts {
		return retryDelayResult{}
	}
	delay, err := platformasync.RetryDelay(attempt, processor.jitter)
	return retryDelayResult{value: delay, err: err}
}

func stableFailureDigest(parts ...string) string {
	sum := sha256.Sum256([]byte("external-job.v1\x00" + strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}
