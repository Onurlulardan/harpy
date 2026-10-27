package transaction

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

func (coordinator *Coordinator) InJobTenant(
	ctx context.Context, tenantID uuid.UUID,
	execute func(context.Context, JobTenantScope) error,
) error {
	return coordinator.inJobTenant(
		ctx, tenantID, postgres.TransactionCommand, execute)
}

// InRepeatableReadJobTenant provides one read-write REPEATABLE READ tenant
// transaction for jobs that must pin a stable source set before committing
// their receipt/state transition. It has the same owner binding fence as a
// normal JobTenantScope and never permits nested scopes.
func (coordinator *Coordinator) InRepeatableReadJobTenant(
	ctx context.Context, tenantID uuid.UUID,
	execute func(context.Context, JobTenantScope) error,
) error {
	return coordinator.inJobTenant(
		ctx, tenantID, postgres.TransactionRepeatableCommand, execute)
}
