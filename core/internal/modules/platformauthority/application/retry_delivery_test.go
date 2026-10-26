package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func TestRetryFailsClosedWithoutStore(t *testing.T) {
	err := (*DeliveryQueue)(nil).RetryDelivery(
		context.Background(), ports.DeliveryClaim{}, "RETRY", false,
	)
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("retry error = %v", err)
	}
}
