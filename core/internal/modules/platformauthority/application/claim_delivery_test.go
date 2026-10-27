package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func TestClaimFailsClosedWithoutStore(t *testing.T) {
	_, _, err := (*DeliveryQueue)(nil).ClaimDelivery(context.Background(), "owner")
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("claim error = %v", err)
	}
}
