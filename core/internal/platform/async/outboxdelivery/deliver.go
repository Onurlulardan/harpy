package outboxdelivery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
)

func (runner *Runner) deliver(
	ctx context.Context,
	broker Broker,
	claim asyncpostgres.PublishClaim,
) error {
	publishContext, cancel := context.WithTimeout(ctx, runner.settings.ConfirmTimeout)
	receipt, publishErr := broker.Publish(publishContext, brokerrabbit.PublishRequest{
		Envelope: claim.Envelope(), LeaseGeneration: claim.LeaseGeneration(),
	})
	cancel()
	if publishErr == nil {
		if _, err := runner.store.ConfirmPublish(ctx, claim, receipt.ConfirmSHA256); err != nil {
			return fmt.Errorf("record outbox publisher confirmation: %w", err)
		}
		return nil
	}
	code, digest, reconnect := publishFailureEvidence(claim, publishErr)
	attempt := claim.DeliveryAttempt()
	if attempt > 25 {
		attempt = 25
	}
	delay, err := platformasync.RetryDelay(attempt, runner.jitter)
	if err != nil {
		return fmt.Errorf("calculate outbox retry delay: %w", err)
	}
	if _, err := runner.store.ReleasePublish(ctx, claim, delay, code, digest); err != nil {
		return fmt.Errorf("release outbox publisher claim: %w", err)
	}
	if reconnect {
		return errors.New("RabbitMQ publisher session requires reconnect")
	}
	return nil
}

func publishFailureEvidence(
	claim asyncpostgres.PublishClaim,
	err error,
) (string, string, bool) {
	var failure *brokerrabbit.PublishFailure
	if errors.As(err, &failure) {
		return failure.Code(), failure.SHA256(), failure.Code() != brokerrabbit.FailureUnroutable
	}
	document := fmt.Sprintf(
		"outbox-publisher-failure.v1\n%s\n%d",
		claim.Envelope().EventID, claim.LeaseGeneration(),
	)
	digest := sha256.Sum256([]byte(document))
	return brokerrabbit.FailurePublish, hex.EncodeToString(digest[:]), true
}
