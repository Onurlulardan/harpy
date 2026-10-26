package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
)

type licenseTargetReader func(
	context.Context,
) (billing.LocalLicenseRefreshTarget, error)

type licenseReconcile func(
	context.Context,
	billing.LocalLicenseRefreshCommand,
) error

type licenseRefreshRunner struct {
	target    licenseTargetReader
	reconcile licenseReconcile
	delivery  string
}

func (runner licenseRefreshRunner) RunOnce(
	ctx context.Context,
) (time.Duration, error) {
	if ctx == nil || ctx.Err() != nil || runner.target == nil ||
		runner.reconcile == nil || runner.delivery == "" {
		return 0, fmt.Errorf("license refresh runner is invalid")
	}
	target, err := runner.target(ctx)
	if err != nil {
		return 0, err
	}
	if !target.Found || !target.Ready {
		return boundedLicenseDelay(target.RetryAt.Sub(target.DatabaseNow)), nil
	}
	if string(target.Binding.DeliveryKind) != runner.delivery {
		return 0, fmt.Errorf("license refresh delivery binding changed")
	}
	err = runner.reconcile(ctx, billing.LocalLicenseRefreshCommand{
		OperationID: target.OperationID,
		Binding:     target.Binding,
	})
	if err == nil {
		return time.Second, nil
	}
	if billing.LocalLicenseRefreshRetryable(err) {
		return 30 * time.Second, nil
	}
	return 0, err
}

func boundedLicenseDelay(value time.Duration) time.Duration {
	if value < time.Second {
		return time.Second
	}
	if value > 5*time.Minute {
		return 5 * time.Minute
	}
	return value
}
