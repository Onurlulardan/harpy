package systemtransfer

import (
	"context"
	"errors"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

const exportReconciliationMaximumRetryDelay = time.Duration(jobsv1.RetryCapMilliseconds) * time.Millisecond

func (processor *ExportReconciliationProcessor) resolveUnknown(
	ctx context.Context,
	claim ExportReconciliationClaim,
	observation ExportReconciliationObservation,
) error {
	now := processor.clock.Now()
	if !exportArtifactTime(now) || now.Before(observation.StartedAt) {
		return ErrExportReconciliationInvalid
	}
	code := ExportReconciliationUnknownCode
	if claim.Attempt >= claim.MaximumAttempts {
		code = ExportReconciliationParkedCode
	}
	record := ExportReconciliationUnknownRecord{Code: code,
		FailureSHA256: exportReconciliationFailureSHA(claim, observation, code),
		ObservedAt:    now}
	if code == ExportReconciliationUnknownCode {
		delay, err := exportReconciliationRetryDelay(processor.backoff, claim.Attempt)
		if err != nil || delay < 0 ||
			delay > exportReconciliationRetryLimit(claim.Attempt) {
			return errors.Join(ErrExportReconciliationInvalid, err)
		}
		record.RetryAt = now.Add(delay).UTC().Truncate(time.Microsecond)
	}
	if record.Validate(claim, observation) != nil {
		return ErrExportReconciliationInvalid
	}
	finalize, cancel := exportReconciliationFinalizeContext(ctx)
	defer cancel()
	var err error
	if code == ExportReconciliationParkedCode {
		err = processor.store.PersistReconciliationParked(
			finalize, claim, observation, record)
	} else {
		err = processor.store.ScheduleReconciliationRetry(
			finalize, claim, observation, record)
	}
	if err != nil {
		return errors.Join(ErrExportReconciliationPersist, err)
	}
	return nil
}

func exportReconciliationRetryLimit(attempt int) time.Duration {
	limit := time.Duration(jobsv1.RetryBaseMilliseconds) * time.Millisecond
	for current := 1; current < attempt && limit < exportReconciliationMaximumRetryDelay; current++ {
		limit *= 2
		if limit > exportReconciliationMaximumRetryDelay {
			return exportReconciliationMaximumRetryDelay
		}
	}
	return limit
}

func exportReconciliationRetryDelay(
	backoff ExportReconciliationBackoff,
	attempt int,
) (delay time.Duration, err error) {
	defer func() {
		if recover() != nil {
			delay = 0
			err = ErrExportUnavailable
		}
	}()
	return backoff.RetryDelay(attempt)
}

func (record ExportReconciliationUnknownRecord) Validate(
	claim ExportReconciliationClaim,
	observation ExportReconciliationObservation,
) error {
	if claim.Validate() != nil || observation.Validate(claim) != nil ||
		record.FailureSHA256 == [32]byte{} || !exportArtifactTime(record.ObservedAt) ||
		record.ObservedAt.Before(observation.StartedAt) {
		return ErrExportReconciliationInvalid
	}
	if record.Code == ExportReconciliationUnknownCode {
		if claim.Attempt >= claim.MaximumAttempts ||
			!exportArtifactTime(record.RetryAt) || record.RetryAt.Before(record.ObservedAt) ||
			record.RetryAt.Sub(record.ObservedAt) >
				exportReconciliationRetryLimit(claim.Attempt) {
			return ErrExportReconciliationInvalid
		}
		return nil
	}
	if record.Code != ExportReconciliationParkedCode ||
		claim.Attempt != claim.MaximumAttempts || !record.RetryAt.IsZero() {
		return ErrExportReconciliationInvalid
	}
	return nil
}
