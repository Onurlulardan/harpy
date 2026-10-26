package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func TestCompleteDeliveryRejectsInvalidLease(t *testing.T) {
	err := (*Store)(nil).CompleteDelivery(
		context.Background(), ports.DeliveryClaim{},
	)
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("complete error = %v", err)
	}
}
