package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func TestUpdateDeliveryImportRejectsInvalidLease(t *testing.T) {
	err := (*Store)(nil).UpdateDeliveryImport(
		context.Background(), ports.DeliveryClaim{},
	)
	if !errors.Is(err, ports.ErrConflict) {
		t.Fatalf("await error = %v", err)
	}
}
