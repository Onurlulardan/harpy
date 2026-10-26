package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
)

type authorityDeliveryRunner struct {
	queue     platformauthority.DeliveryQueue
	processor platformauthority.DeliveryProcessor
	ownerID   string
}

func (runner authorityDeliveryRunner) RunOnce(
	ctx context.Context,
) (time.Duration, error) {
	if ctx == nil || ctx.Err() != nil || runner.queue == nil ||
		runner.processor == nil || runner.ownerID == "" {
		return 0, fmt.Errorf("platform authority delivery runner is invalid")
	}
	claim, found, err := runner.queue.ClaimDelivery(ctx, runner.ownerID)
	if err != nil {
		return 0, err
	}
	if !found {
		return 2 * time.Second, nil
	}
	err = runner.processor.Process(ctx, claim.OperationID)
	if errors.Is(err, platformauthority.ErrDeliveryAwaitingImport) {
		if dispositionErr := runner.queue.UpdateDeliveryImport(
			ctx, claim,
		); dispositionErr != nil {
			return 0, dispositionErr
		}
		return time.Second, nil
	}
	if err != nil {
		permanent := errors.Is(err, platformauthority.ErrDeliveryPermanent) ||
			errors.Is(err, platformauthority.ErrDeliveryFinalizerUnavailable)
		code := "DELIVERY_RETRY"
		if permanent {
			code = "DELIVERY_INVALID"
		}
		if dispositionErr := runner.queue.RetryDelivery(
			ctx, claim, code, permanent,
		); dispositionErr != nil {
			return 0, dispositionErr
		}
		return time.Second, nil
	}
	if err := runner.queue.CompleteDelivery(ctx, claim); err != nil {
		return 0, err
	}
	return time.Second, nil
}
