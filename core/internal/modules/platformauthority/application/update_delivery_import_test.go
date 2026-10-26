package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func TestUpdateFailsClosedWithoutStore(t *testing.T) {
	err := (*DeliveryQueue)(nil).UpdateDeliveryImport(
		context.Background(), ports.DeliveryClaim{},
	)
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("await error = %v", err)
	}
}
