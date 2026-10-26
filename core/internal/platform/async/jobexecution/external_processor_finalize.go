package jobexecution

import (
	"context"
	"errors"
	"fmt"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func (processor *ExternalProcessor) finalizeResult(
	ctx context.Context,
	claim asyncpostgres.JobClaim,
	attempt int,
	started StartedEffect,
	result effectexecution.Result,
) error {
	if result.IsUnknown() {
		return processor.markUnknown(ctx, started)
	}
	receipt, present := result.Receipt()
	if !present {
		return processor.markUnknown(ctx, started)
	}
	finalizeContext, cancel := boundedFinalizationContext(ctx)
	defer cancel()
	switch receipt.Status {
	case effectsv1.ReceiptSucceeded:
		if _, err := processor.finalizer.CompleteExternalJob(
			finalizeContext, started.permit, receipt,
		); err != nil {
			return fmt.Errorf("complete external job: %w", err)
		}
		return nil
	case effectsv1.ReceiptFailed:
		failure, err := processor.failureFromReceipt(receipt, attempt)
		if err != nil {
			return errorsAfterUnknown(processor, ctx, started, err)
		}
		if _, err := processor.finalizer.FailExternalJob(
			finalizeContext, started.permit, receipt, failure,
		); err != nil {
			return fmt.Errorf("fail external job: %w", err)
		}
		return nil
	default:
		return processor.markUnknown(ctx, started)
	}
}

func (processor *ExternalProcessor) markUnknown(
	ctx context.Context,
	started StartedEffect,
) error {
	finalizeContext, cancel := boundedFinalizationContext(ctx)
	defer cancel()
	if _, err := processor.finalizer.MarkExternalJobOutcomeUnknown(
		finalizeContext, started.permit,
	); err != nil {
		return fmt.Errorf("mark external job outcome unknown: %w", err)
	}
	return nil
}

func boundedFinalizationContext(ctx context.Context) (context.Context, context.CancelFunc) {
	parent := context.Background()
	if ctx != nil {
		parent = context.WithoutCancel(ctx)
	}
	return context.WithTimeout(parent, externalFinalizationTimeout)
}

func errorsAfterUnknown(
	processor *ExternalProcessor,
	ctx context.Context,
	started StartedEffect,
	cause error,
) error {
	if unknownErr := processor.markUnknown(ctx, started); unknownErr != nil {
		return errors.Join(fmt.Errorf("invalid failed effect receipt: %w", cause), unknownErr)
	}
	return nil
}
