package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func TestEmptyDeliveryClaimIsNotUsable(t *testing.T) {
	if usableDeliveryClaim(ports.DeliveryClaim{}) || usableDeliveryClaim(ports.DeliveryClaim{
		OperationID: "operation", LeaseGeneration: 0,
	}) {
		t.Fatal("empty delivery claim was treated as live work")
	}
	if !usableDeliveryClaim(ports.DeliveryClaim{
		OperationID: "operation", LeaseGeneration: 1,
	}) {
		t.Fatal("leased delivery claim was ignored")
	}
}

func TestClaimDeliveryRejectsEmptyOwner(t *testing.T) {
	_, _, err := (*Store)(nil).ClaimDelivery(context.Background(), "")
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("claim error = %v", err)
	}
}
