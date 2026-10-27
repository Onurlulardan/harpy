package billingv1

import (
	"context"
	"fmt"
)

func (handler *Handler) Admit(ctx context.Context) error {
	if handler == nil || ctx == nil || ctx.Err() != nil {
		return fmt.Errorf("Billing V1 compatibility admission is invalid")
	}
	handler.admitted.Store(false)
	decision, err := handler.retirement.AdmitManifest(ctx, handler.manifest)
	if err != nil || decision.Validate() != nil {
		if err != nil {
			return fmt.Errorf("admit Billing V1 route manifest: %w", err)
		}
		return fmt.Errorf("admit Billing V1 route manifest: invalid decision")
	}
	retired, err := decision.RetiredCount()
	if err != nil || retired != 0 {
		return fmt.Errorf("admit Billing V1 route manifest: retired route")
	}
	handler.admitted.Store(true)
	return nil
}
