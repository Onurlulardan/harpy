package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func TestCompleteFailsClosedWithoutStore(t *testing.T) {
	err := (*DeliveryQueue)(nil).CompleteDelivery(
		context.Background(), ports.DeliveryClaim{},
	)
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("complete error = %v", err)
	}
}
