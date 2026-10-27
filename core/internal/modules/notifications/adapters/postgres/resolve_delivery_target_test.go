package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func TestResolveDeliveryTargetRejectsInvalidRuntime(t *testing.T) {
	var runtime *DeliveryRuntime
	if _, err := runtime.DeliveryTarget(t.Context(), domain.ID{}, domain.ID{}, ""); err != ErrInvalid {
		t.Fatalf("DeliveryTarget() error = %v", err)
	}
}
