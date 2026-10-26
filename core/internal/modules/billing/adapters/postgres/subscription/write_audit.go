package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) writeAudit(
	ctx context.Context,
	authorityID uuid.UUID,
	sourceKind string,
	plan subscriptionapp.CommitPlan,
) error {
	var actor any
	if sourceKind == string(modeActor) {
		actor = plan.Audit.ActorID
	}
	err := storage.expectID(
		ctx, insertAuditSQL, plan.Audit.ID,
		authorityID, storage.tenantID, plan.Audit.ID,
		plan.Audit.SubscriptionID, plan.Audit.SourceOperationID,
		sourceKind, plan.Audit.Kind, actor, plan.Audit.BeforeRevision,
		plan.Audit.AfterRevision, plan.Audit.OccurredAt, plan.Audit.Reason,
	)
	if err != nil {
		return fmt.Errorf("insert billing subscription audit: %w", err)
	}
	return nil
}
