package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func TestClaimDeliveryRejectsEmptyOwner(t *testing.T) {
	_, _, err := (*Store)(nil).ClaimDelivery(context.Background(), "")
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("claim error = %v", err)
	}
}
