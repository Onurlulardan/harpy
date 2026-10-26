package worker

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/deadletterdelivery"
)

func probeDeadLetterDelivery(
	ctx context.Context,
	connector deadletterdelivery.Connector,
) error {
	if ctx == nil || connector == nil {
		return fmt.Errorf("dead-letter dependency preflight is incomplete")
	}
	session, err := connector.Open(ctx)
	if err != nil {
		return fmt.Errorf("dead-letter dependency preflight: %w", err)
	}
	if session == nil {
		return fmt.Errorf("dead-letter dependency preflight returned no session")
	}
	if err := session.Close(); err != nil {
		return fmt.Errorf("close dead-letter dependency preflight: %w", err)
	}
	return nil
}
