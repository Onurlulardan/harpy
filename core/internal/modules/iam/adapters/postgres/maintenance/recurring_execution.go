package maintenance

import (
	"context"
	"fmt"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	"github.com/google/uuid"
)

func (store *store) executeSimple(
	ctx context.Context, claim maintenanceapp.Claim, nextID uuid.UUID, query string,
) (maintenanceapp.Outcome, error) {
	if ctx == nil || ctx.Err() != nil || store == nil || !store.runtime.usable() {
		return maintenanceapp.Outcome{}, ErrRuntimeInvalid
	}
	values, err := claimValues(claim)
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	if receipt, found, err := store.readReceipt(ctx, claim); found || err != nil {
		return receipt, err
	}
	values = append(values, nextID)
	outcome, err := scanOutcome(store.database.QueryRow(ctx, query, values...), claim.Kind)
	if err != nil {
		return maintenanceapp.Outcome{}, fmt.Errorf("execute %s: %w", claim.Kind, err)
	}
	return outcome, nil
}
