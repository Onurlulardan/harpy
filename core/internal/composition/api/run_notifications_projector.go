package api

import (
	"context"
	"fmt"
)

func (workload *notificationsProjectorWorkload) run(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return nil
		}
		if err := workload.drainActiveTenants(ctx); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if err := workload.wait(ctx, notificationsProjectionInterval); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("wait for notification projection: %w", err)
		}
	}
}

func (workload *notificationsProjectorWorkload) drainActiveTenants(ctx context.Context) error {
	for _, tenantID := range workload.realtime.ActiveTenantIDs() {
		if tenantID == "" {
			return fmt.Errorf("notification projection tenant is invalid")
		}
		if err := workload.drainTenant(ctx, tenantID); err != nil {
			return err
		}
	}
	return nil
}

func (workload *notificationsProjectorWorkload) drainTenant(
	ctx context.Context, tenantID string,
) error {
	for batch := 0; batch < notificationsProjectionMaxBatches; batch++ {
		count, err := workload.projector.DrainTenant(ctx,
			notificationsProjectionActorID, tenantID, notificationsProjectionBatchSize)
		if err != nil {
			return fmt.Errorf("project notification realtime hints: %w", err)
		}
		if count < notificationsProjectionBatchSize {
			return nil
		}
	}
	return nil
}
