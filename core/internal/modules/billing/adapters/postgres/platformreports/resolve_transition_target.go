package platformreports

import (
	"context"

	platformreportapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformreports"
	"github.com/google/uuid"
)

type row interface{ Scan(...any) error }
type queryRow func(context.Context, string, ...any) row

func (storage *store) ResolveTransitionTarget(
	ctx context.Context,
	subscriptionID uuid.UUID,
) (platformreportapp.TransitionTarget, error) {
	if storage == nil || storage.query == nil || ctx == nil || ctx.Err() != nil ||
		subscriptionID == uuid.Nil {
		return platformreportapp.TransitionTarget{}, ErrConfiguration
	}
	var target platformreportapp.TransitionTarget
	err := storage.query(ctx, resolveTransitionTargetSQL, subscriptionID).Scan(
		&target.TenantID, &target.Provider, &target.GroupGeneration,
		&target.BillingGeneration, &target.FormsGeneration,
	)
	if err != nil || !target.Valid() {
		return platformreportapp.TransitionTarget{}, platformreportapp.ErrTarget
	}
	return target, nil
}

var _ platformreportapp.TransitionTargetStore = (*store)(nil)
