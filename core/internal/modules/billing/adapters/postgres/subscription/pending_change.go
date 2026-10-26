package subscription

import (
	"context"
	"fmt"
	"time"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) findPendingChange(
	ctx context.Context,
	authorityID, subscriptionID uuid.UUID,
) (*subscriptionapp.SubscriptionChange, error) {
	var change subscriptionapp.SubscriptionChange
	var state string
	var effectiveAt, createdAt time.Time
	err := storage.queryRow(
		ctx, findPendingChangeSQL, authorityID, storage.tenantID, subscriptionID,
	).Scan(
		&change.ID, &change.DecisionID, &change.SubscriptionID,
		&change.BeforeOfferingVersionID, &change.BeforePriceVersionID,
		&change.AfterOfferingVersionID, &change.AfterPriceVersionID,
		&change.SourceOperationID, &state, &effectiveAt, &createdAt, &change.Revision,
	)
	if err != nil {
		if storage.notFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("find pending billing subscription change: %w", err)
	}
	change.State = subscriptionapp.ChangeState(state)
	change.EffectiveAt, change.CreatedAt = effectiveAt.UTC(), createdAt.UTC()
	return &change, nil
}
