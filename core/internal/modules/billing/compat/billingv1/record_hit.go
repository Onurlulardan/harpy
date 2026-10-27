package billingv1

import (
	"context"
	"fmt"
)

func (handler *Handler) recordHit(ctx context.Context, routeID string) error {
	if handler == nil || ctx == nil || !handler.admitted.Load() {
		return fmt.Errorf("Billing V1 compatibility is not admitted")
	}
	batch, found := handler.keys[routeID]
	if !found {
		return fmt.Errorf("Billing V1 route evidence is invalid")
	}
	decision, err := handler.retirement.RecordHits(ctx, batch)
	if err != nil || decision.Validate() != nil {
		if err != nil {
			return fmt.Errorf("record Billing V1 route hit: %w", err)
		}
		return fmt.Errorf("record Billing V1 route hit: invalid decision")
	}
	retired, err := decision.RetiredCount()
	if err != nil || retired != 0 {
		return fmt.Errorf("record Billing V1 route hit: retired route")
	}
	return nil
}
