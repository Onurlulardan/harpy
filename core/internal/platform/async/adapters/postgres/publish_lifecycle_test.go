package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPublishLifecycleRejectsInvalidMutationBeforeTransaction(t *testing.T) {
	t.Parallel()
	database := &workerTestDatabase{}
	repository := newWorkerTestRepository()
	repository.database = database
	claim := validPublishClaim(t, repository)
	digest := strings.Repeat("a", 64)
	if _, err := repository.ConfirmPublish(
		context.Background(), claim, "BAD",
	); !errors.Is(err, ErrPublishConfirmDigest) {
		t.Fatalf("confirm error = %v", err)
	}
	invalidReleases := []struct {
		delay time.Duration
		code  string
		hash  string
	}{
		{delay: -time.Microsecond, code: "BROKER_NACK", hash: digest},
		{delay: maximumPublishRetryDelay + time.Microsecond, code: "BROKER_NACK", hash: digest},
		{delay: time.Nanosecond, code: "BROKER_NACK", hash: digest},
		{code: "bad", hash: digest},
		{code: "BROKER_NACK", hash: "BAD"},
	}
	for _, invalid := range invalidReleases {
		if _, err := repository.ReleasePublish(
			context.Background(), claim, invalid.delay, invalid.code, invalid.hash,
		); !errors.Is(err, ErrPublishReleaseInvalid) {
			t.Fatalf("release error = %v", err)
		}
	}
	for _, limit := range []int{0, 51} {
		if _, err := repository.RecoverExpiredOutbox(
			context.Background(), claim.authority, limit,
		); !errors.Is(err, ErrPublishRecoveryLimit) {
			t.Fatalf("recovery limit %d error = %v", limit, err)
		}
		if _, err := repository.ParkStaleOutbox(
			context.Background(), claim.authority, limit,
		); !errors.Is(err, ErrPublishRecoveryLimit) {
			t.Fatalf("stale park limit %d error = %v", limit, err)
		}
	}
	if database.beginCalls != 0 {
		t.Fatalf("invalid mutations began %d transactions", database.beginCalls)
	}
}

func TestPublishLifecycleUsesAuthorityTransactionForValidInputs(t *testing.T) {
	t.Parallel()
	database := &workerTestDatabase{}
	repository := newWorkerTestRepository()
	repository.database = database
	claim := validPublishClaim(t, repository)
	digest := strings.Repeat("a", 64)
	operations := []func() error{
		func() error { _, err := repository.HeartbeatPublish(context.Background(), claim); return err },
		func() error { _, err := repository.ConfirmPublish(context.Background(), claim, digest); return err },
		func() error {
			_, err := repository.ReleasePublish(
				context.Background(), claim, maximumPublishRetryDelay,
				"BROKER_NACK", digest,
			)
			return err
		},
		func() error {
			_, err := repository.RecoverExpiredOutbox(context.Background(), claim.authority, 50)
			return err
		},
		func() error {
			_, err := repository.ParkStaleOutbox(context.Background(), claim.authority, 50)
			return err
		},
	}
	for _, operation := range operations {
		if err := operation(); !errors.Is(err, errWorkerTestDatabase) {
			t.Fatalf("operation error = %v", err)
		}
	}
	if database.beginCalls != len(operations) {
		t.Fatalf("valid mutations began %d transactions", database.beginCalls)
	}
}
