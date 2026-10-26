package worker

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
)

func TestLicenseRefreshRunnerKeepsRetryableAndTerminalOutcomesDistinct(t *testing.T) {
	for _, delivery := range []billing.LocalLicenseDeliveryKind{
		billing.LocalLicenseDeliveryConnected,
		billing.LocalLicenseDeliveryAirGap,
		billing.LocalLicenseDeliveryDevelopment,
	} {
		t.Run(string(delivery), func(t *testing.T) {
			target := licenseTarget(t, delivery)
			retryable := licenseRefreshRunner{
				target: func(context.Context) (billing.LocalLicenseRefreshTarget, error) {
					return target, nil
				},
				reconcile: func(
					context.Context,
					billing.LocalLicenseRefreshCommand,
				) error {
					return billing.ErrLocalLicenseRefreshRetryable
				},
				delivery: string(delivery),
			}
			delay, err := retryable.RunOnce(context.Background())
			if err != nil || delay != 30*time.Second {
				t.Fatalf("retryable delay=%s err=%v", delay, err)
			}
			terminal := retryable
			terminal.reconcile = func(
				context.Context,
				billing.LocalLicenseRefreshCommand,
			) error {
				return billing.ErrLocalLicenseRefreshFenced
			}
			if _, err = terminal.RunOnce(context.Background()); !errors.Is(
				err,
				billing.ErrLocalLicenseRefreshFenced,
			) {
				t.Fatalf("terminal error=%v", err)
			}
		})
	}
}

func TestLicenseRefreshRunnerRejectsCrossDeliveryAndBoundsIdle(t *testing.T) {
	target := licenseTarget(t, billing.LocalLicenseDeliveryAirGap)
	target.Ready = false
	target.RetryAt = target.DatabaseNow.Add(time.Hour)
	runner := licenseRefreshRunner{
		target: func(context.Context) (billing.LocalLicenseRefreshTarget, error) {
			return target, nil
		},
		reconcile: func(
			context.Context,
			billing.LocalLicenseRefreshCommand,
		) error {
			return nil
		},
		delivery: string(billing.LocalLicenseDeliveryConnected),
	}
	delay, err := runner.RunOnce(context.Background())
	if err != nil || delay != 5*time.Minute {
		t.Fatalf("idle delay=%s err=%v", delay, err)
	}
	target.Ready, target.RetryAt = true, target.DatabaseNow.Add(time.Minute)
	runner.target = func(context.Context) (billing.LocalLicenseRefreshTarget, error) {
		return target, nil
	}
	if _, err = runner.RunOnce(context.Background()); err == nil || !strings.Contains(err.Error(), "delivery binding") {
		t.Fatalf("cross-delivery error=%v", err)
	}
}

func licenseTarget(
	t *testing.T,
	delivery billing.LocalLicenseDeliveryKind,
) billing.LocalLicenseRefreshTarget {
	t.Helper()
	operation, err := billing.NewUUIDv7()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	return billing.LocalLicenseRefreshTarget{
		Found: true, Ready: true, OperationID: operation,
		Binding: billing.LocalLicenseBinding{
			LicenseID:              billing.NewUUID(),
			LicenseeOrganizationID: billing.NewUUID(),
			EnvironmentID:          billing.NewUUID(), DeliveryKind: delivery,
		},
		Revision: 2, DatabaseNow: now, RetryAt: now.Add(time.Minute),
	}
}
