package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func TestRetryDeliveryRejectsInvalidLease(t *testing.T) {
	err := (*Store)(nil).RetryDelivery(
		context.Background(), ports.DeliveryClaim{}, "RETRY", false,
	)
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("retry error = %v", err)
	}
}
