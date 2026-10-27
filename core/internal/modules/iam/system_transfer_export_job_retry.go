package iam

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

func (processor *systemTransferExportJobProcessor) persistPreEffectResult(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
	result systemtransfer.ExportExecutionResult,
	cause error,
) error {
	class := jobsv1.ErrorPermanent
	if result.State == systemtransfer.ExportExecutionRetry ||
		result.Code == systemtransfer.ExportCodeDependencyUnavailable {
		class = jobsv1.ErrorTransient
	}
	if result.State == systemtransfer.ExportExecutionCancelled ||
		errors.Is(cause, systemtransfer.ErrExportCancelled) {
		class = jobsv1.ErrorCancelled
	}
	delay := time.Duration(0)
	var err error
	if class == jobsv1.ErrorTransient && claim.Attempt() < claim.MaximumAttempts() {
		delay, err = platformasync.RetryDelay(claim.Attempt(), systemTransferImportJitter{})
		if err != nil {
			return fmt.Errorf("calculate IAM export retry: %w", err)
		}
	}
	code := result.Code
	if code == "" {
		code = systemtransfer.ExportCodeDependencyUnavailable
	}
	failure, err := asyncpostgres.NewJobFailure(
		class, code, systemTransferExportFailureDigest(code, cause), delay,
	)
	if err != nil {
		return fmt.Errorf("build IAM export pre-effect failure: %w", err)
	}
	finalizeContext, cancel := context.WithTimeout(
		context.WithoutCancel(ctx), systemTransferExportFinalizationTimeout,
	)
	defer cancel()
	settled, err := processor.preEffect.FailBeforeEffect(
		finalizeContext, claim, failure,
	)
	if err != nil {
		return fmt.Errorf("persist IAM export pre-effect failure: %w", err)
	}
	switch class {
	case jobsv1.ErrorTransient:
		if claim.Attempt() < claim.MaximumAttempts() {
			if _, queued := settled.RetryAt(); !queued {
				return ErrSystemTransferJobPolicy
			}
			return nil
		}
		if settled.State() != jobsv1.StateDead {
			return ErrSystemTransferJobPolicy
		}
	case jobsv1.ErrorCancelled:
		if settled.State() != jobsv1.StateCancelled {
			return ErrSystemTransferJobPolicy
		}
	case jobsv1.ErrorPermanent:
		if settled.State() != jobsv1.StateDead {
			return ErrSystemTransferJobPolicy
		}
	}
	return nil
}

func systemTransferExportFailureDigest(code string, cause error) string {
	message := ""
	if cause != nil {
		message = cause.Error()
	}
	digest := sha256.Sum256([]byte("iam-system-export.v1\x00" + code + "\x00" + message))
	return hex.EncodeToString(digest[:])
}
